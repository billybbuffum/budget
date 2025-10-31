#!/bin/bash

# Script to run the AllocateToCoverUnderfunded tests

set -e

echo "=========================================="
echo "Running AllocateToCoverUnderfunded Tests"
echo "=========================================="
echo ""

echo "1. Running AllocationService unit tests..."
echo "-------------------------------------------"
go test ./internal/application -v -run "TestAllocationService_(AllocateToCoverUnderfunded|CalculateReadyToAssignWithoutUnderfunded)"
echo ""

echo "2. Running AllocationHandler integration tests..."
echo "--------------------------------------------------"
go test ./internal/infrastructure/http/handlers -v -run "TestAllocationHandler_CoverUnderfunded"
echo ""

echo "3. Running all tests with coverage..."
echo "-------------------------------------"
go test ./internal/application -coverprofile=app_coverage.out
go test ./internal/infrastructure/http/handlers -coverprofile=handler_coverage.out
echo ""

echo "4. Coverage Summary:"
echo "--------------------"
echo "Application Layer:"
go tool cover -func=app_coverage.out | grep -E "(AllocateToCoverUnderfunded|calculateReadyToAssignWithoutUnderfunded)"
echo ""
echo "Handler Layer:"
go tool cover -func=handler_coverage.out | grep "CoverUnderfunded"
echo ""

echo "=========================================="
echo "All tests completed successfully!"
echo "=========================================="
echo ""
echo "To view detailed coverage:"
echo "  go tool cover -html=app_coverage.out"
echo "  go tool cover -html=handler_coverage.out"
