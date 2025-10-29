package ofx

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aclindsa/ofxgo"
)

// ParsedTransaction represents a transaction parsed from an OFX file
type ParsedTransaction struct {
	Date        time.Time
	Amount      int64  // In cents
	Description string
	FitID       string // Financial institution transaction ID (for duplicate detection)
}

// ImportResult contains the result of parsing an OFX file
type ImportResult struct {
	Transactions  []ParsedTransaction
	AccountID     string // OFX account ID
	Currency      string
	LedgerBalance int64  // Current balance from OFX file (in cents), 0 if not available
}

// Parser handles OFX file parsing
type Parser struct{}

// NewParser creates a new OFX parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses an OFX file and extracts transaction data
// Only imports transactions from the last 90 days to avoid processing years of historical data
func (p *Parser) Parse(reader io.Reader) (*ImportResult, error) {
	// Preprocess the file to handle non-standard line endings (e.g., OnPoint's \r\r\n)
	preprocessed, err := p.preprocessOFX(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess OFX file: %w", err)
	}

	// Parse the OFX response
	response, err := ofxgo.ParseResponse(preprocessed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OFX file: %w", err)
	}

	result := &ImportResult{
		Transactions: []ParsedTransaction{},
	}

	// Calculate cutoff date (90 days ago)
	cutoffDate := time.Now().AddDate(0, 0, -90)

	// Process banking statements
	if len(response.Bank) > 0 {
		for _, msg := range response.Bank {
			if stmt, ok := msg.(*ofxgo.StatementResponse); ok {
				if err := p.processBankStatement(stmt, result, cutoffDate); err != nil {
					return nil, err
				}
			}
		}
	}

	// Process credit card statements
	if len(response.CreditCard) > 0 {
		for _, msg := range response.CreditCard {
			if stmt, ok := msg.(*ofxgo.CCStatementResponse); ok {
				if err := p.processCreditCardStatement(stmt, result, cutoffDate); err != nil {
					return nil, err
				}
			}
		}
	}

	// Note: We allow zero transactions since we're primarily interested in the ledger balance
	// Transactions are optional and only used for categorization

	return result, nil
}

// processBankStatement processes a bank statement from OFX
func (p *Parser) processBankStatement(stmt *ofxgo.StatementResponse, result *ImportResult, cutoffDate time.Time) error {
	// Set account ID
	result.AccountID = string(stmt.BankAcctFrom.AcctID)

	// Set currency
	if valid, _ := stmt.CurDef.Valid(); valid {
		result.Currency = stmt.CurDef.String()
	}

	// Extract ledger balance if available
	if stmt.BalAmt.Rat.Sign() != 0 {
		balanceFloat := stmt.BalAmt.Rat.FloatString(2)
		var balanceVal float64
		fmt.Sscanf(balanceFloat, "%f", &balanceVal)
		result.LedgerBalance = int64(balanceVal * 100)
	}

	// Process transactions (only last 90 days)
	txList := stmt.BankTranList
	if txList == nil {
		return nil
	}

	for _, txn := range txList.Transactions {
		// Skip transactions older than cutoff date
		if txn.DtPosted.Time.Before(cutoffDate) {
			continue
		}

		parsed, err := p.parseTransaction(txn)
		if err != nil {
			// Log error but continue processing other transactions
			continue
		}
		result.Transactions = append(result.Transactions, *parsed)
	}

	return nil
}

// processCreditCardStatement processes a credit card statement from OFX
func (p *Parser) processCreditCardStatement(stmt *ofxgo.CCStatementResponse, result *ImportResult, cutoffDate time.Time) error {
	// Set account ID
	result.AccountID = string(stmt.CCAcctFrom.AcctID)

	// Set currency
	if valid, _ := stmt.CurDef.Valid(); valid {
		result.Currency = stmt.CurDef.String()
	}

	// Extract ledger balance if available
	if stmt.BalAmt.Rat.Sign() != 0 {
		balanceFloat := stmt.BalAmt.Rat.FloatString(2)
		var balanceVal float64
		fmt.Sscanf(balanceFloat, "%f", &balanceVal)
		result.LedgerBalance = int64(balanceVal * 100)
	}

	// Process transactions (only last 90 days)
	txList := stmt.BankTranList
	if txList == nil {
		return nil
	}

	for _, txn := range txList.Transactions {
		// Skip transactions older than cutoff date
		if txn.DtPosted.Time.Before(cutoffDate) {
			continue
		}

		parsed, err := p.parseTransaction(txn)
		if err != nil {
			// Log error but continue processing other transactions
			continue
		}
		result.Transactions = append(result.Transactions, *parsed)
	}

	return nil
}

// parseTransaction converts an OFX transaction to our internal format
func (p *Parser) parseTransaction(txn ofxgo.Transaction) (*ParsedTransaction, error) {
	// Parse date
	date := txn.DtPosted.Time

	// Parse amount - convert from dollars to cents
	// OFX amounts are in dollars as floating point
	amountFloat := txn.TrnAmt.Rat.FloatString(2)
	var amountVal float64
	fmt.Sscanf(amountFloat, "%f", &amountVal)
	amountCents := int64(amountVal * 100)

	// Build description from Name and Memo
	description := p.buildDescription(txn)

	// Get FiTID for duplicate detection
	fitID := string(txn.FiTID)

	return &ParsedTransaction{
		Date:        date,
		Amount:      amountCents,
		Description: description,
		FitID:       fitID,
	}, nil
}

// buildDescription creates a transaction description from OFX Name and Memo fields
func (p *Parser) buildDescription(txn ofxgo.Transaction) string {
	name := strings.TrimSpace(string(txn.Name))
	memo := strings.TrimSpace(string(txn.Memo))

	// If both exist and are different, combine them
	if name != "" && memo != "" && name != memo {
		return fmt.Sprintf("%s - %s", name, memo)
	}

	// If only name exists
	if name != "" {
		return name
	}

	// If only memo exists
	if memo != "" {
		return memo
	}

	// If neither exists, use transaction type
	trnType := string(txn.TrnType)
	if trnType != "" {
		return trnType
	}

	return "Unknown Transaction"
}

// preprocessOFX normalizes line endings and formatting in OFX files
// Different institutions use various non-standard formatting:
// - OnPoint: \r\r\n line endings, tabs before XML, extra blank lines
// - Chase: blank line before headers, mixed line endings
// This function normalizes all variations to proper OFX SGML format
func (p *Parser) preprocessOFX(reader io.Reader) (io.Reader, error) {
	// Read the entire file
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read OFX file: %w", err)
	}

	// Normalize all line ending variations to \n for processing
	// Handle: \r\r\n -> \n, \r\n -> \n, \r -> \n
	normalized := bytes.ReplaceAll(data, []byte("\r\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))

	// Split into lines
	lines := bytes.Split(normalized, []byte("\n"))

	// Find where headers start (skip any leading blank lines)
	headerStartIndex := -1
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 && bytes.Contains(trimmed, []byte("OFXHEADER:")) {
			headerStartIndex = i
			break
		}
	}

	if headerStartIndex == -1 {
		// No OFXHEADER found, might be XML format - return as-is with \r\n line endings
		withCRLF := bytes.ReplaceAll(normalized, []byte("\n"), []byte("\r\n"))
		return bytes.NewReader(withCRLF), nil
	}

	// Find where XML content starts
	xmlStartIndex := -1
	for i := headerStartIndex; i < len(lines); i++ {
		trimmed := bytes.TrimSpace(lines[i])
		if len(trimmed) > 0 && trimmed[0] == '<' {
			xmlStartIndex = i
			break
		}
	}

	if xmlStartIndex == -1 {
		// No XML found, return normalized with \r\n
		withCRLF := bytes.ReplaceAll(normalized, []byte("\n"), []byte("\r\n"))
		return bytes.NewReader(withCRLF), nil
	}

	// Build properly formatted OFX SGML file
	var result [][]byte

	// Valid OFX SGML headers (only these 9 are recognized by the spec)
	validHeaders := []string{
		"OFXHEADER:",
		"DATA:",
		"VERSION:",
		"SECURITY:",
		"ENCODING:",
		"CHARSET:",
		"COMPRESSION:",
		"OLDFILEUID:",
		"NEWFILEUID:",
	}

	// Add header lines (skip blank lines and invalid headers)
	for i := headerStartIndex; i < xmlStartIndex; i++ {
		line := bytes.TrimSpace(lines[i])
		if len(line) == 0 {
			continue
		}

		// Check if this line starts with a valid header
		isValid := false
		for _, validHeader := range validHeaders {
			if bytes.HasPrefix(line, []byte(validHeader)) {
				isValid = true
				break
			}
		}

		if isValid {
			result = append(result, line)
		}
	}

	// Add single blank line after headers (required by SGML spec)
	result = append(result, []byte(""))

	// Add XML content with whitespace trimmed from tags
	for i := xmlStartIndex; i < len(lines); i++ {
		trimmed := bytes.TrimSpace(lines[i])
		if len(trimmed) > 0 {
			result = append(result, trimmed)
		}
	}

	// Join with \r\n (OFX SGML spec requires \r\n)
	cleaned := bytes.Join(result, []byte("\r\n"))
	return bytes.NewReader(cleaned), nil
}

// ValidateOFXFile checks if a file is a valid OFX file
func (p *Parser) ValidateOFXFile(reader io.Reader) error {
	preprocessed, err := p.preprocessOFX(reader)
	if err != nil {
		return fmt.Errorf("failed to preprocess OFX file: %w", err)
	}

	_, err = ofxgo.ParseResponse(preprocessed)
	if err != nil {
		return fmt.Errorf("invalid OFX file: %w", err)
	}
	return nil
}
