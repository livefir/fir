#!/bin/bash

# Test script to run individual action tests
cd /Users/adnaan/code/livefir/fir/examples/e2e

echo "Running AppendAction test..."
go test -run TestActionsE2E/AppendAction -v .

echo ""
echo "Running PrependAction test..."
go test -run TestActionsE2E/PrependAction -v .

echo ""
echo "Running all RefreshAction, ResetAction, ToggleDisabledAction tests..."
go test -run "TestActionsE2E/(RefreshAction|ResetAction|ToggleDisabledAction)" -v .
