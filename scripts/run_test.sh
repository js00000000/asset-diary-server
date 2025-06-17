#!/bin/bash

# Set environment variables
export ENV=development

# Run the test script
echo "Running daily asset job test..."
go run scripts/test_daily_asset_job.go

echo "Test completed."
