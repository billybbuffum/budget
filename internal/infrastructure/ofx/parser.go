package ofx

import (
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
	Transactions []ParsedTransaction
	AccountID    string // OFX account ID
	Currency     string
}

// Parser handles OFX file parsing
type Parser struct{}

// NewParser creates a new OFX parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses an OFX file and extracts transaction data
func (p *Parser) Parse(reader io.Reader) (*ImportResult, error) {
	// Parse the OFX response
	response, err := ofxgo.ParseResponse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OFX file: %w", err)
	}

	result := &ImportResult{
		Transactions: []ParsedTransaction{},
	}

	// Process banking statements
	if len(response.Bank) > 0 {
		for _, msg := range response.Bank {
			if stmt, ok := msg.(*ofxgo.StatementResponse); ok {
				if err := p.processBankStatement(stmt, result); err != nil {
					return nil, err
				}
			}
		}
	}

	// Process credit card statements
	if len(response.CreditCard) > 0 {
		for _, msg := range response.CreditCard {
			if stmt, ok := msg.(*ofxgo.CCStatementResponse); ok {
				if err := p.processCreditCardStatement(stmt, result); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(result.Transactions) == 0 {
		return nil, fmt.Errorf("no transactions found in OFX file")
	}

	return result, nil
}

// processBankStatement processes a bank statement from OFX
func (p *Parser) processBankStatement(stmt *ofxgo.StatementResponse, result *ImportResult) error {
	// Set account ID
	result.AccountID = string(stmt.BankAcctFrom.AcctID)

	// Set currency
	if valid, _ := stmt.CurDef.Valid(); valid {
		result.Currency = stmt.CurDef.String()
	}

	// Process transactions
	txList := stmt.BankTranList
	if txList == nil {
		return nil
	}

	for _, txn := range txList.Transactions {
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
func (p *Parser) processCreditCardStatement(stmt *ofxgo.CCStatementResponse, result *ImportResult) error {
	// Set account ID
	result.AccountID = string(stmt.CCAcctFrom.AcctID)

	// Set currency
	if valid, _ := stmt.CurDef.Valid(); valid {
		result.Currency = stmt.CurDef.String()
	}

	// Process transactions
	txList := stmt.BankTranList
	if txList == nil {
		return nil
	}

	for _, txn := range txList.Transactions {
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

// ValidateOFXFile checks if a file is a valid OFX file
func (p *Parser) ValidateOFXFile(reader io.Reader) error {
	_, err := ofxgo.ParseResponse(reader)
	if err != nil {
		return fmt.Errorf("invalid OFX file: %w", err)
	}
	return nil
}
