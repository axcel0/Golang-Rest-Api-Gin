#!/bin/bash

################################################################################
# 🚀 Quick Test Runner with Server Auto-Detection
################################################################################

echo "🔍 Checking if server is running..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ Server is NOT running!"
    echo ""
    echo "Please start the server first:"
    echo "  cd cmd/api && go run main.go"
    echo ""
    exit 1
fi

echo "✅ Server is running!"
echo ""
echo "🧪 Running comprehensive test suite..."
echo ""

# Run the test suite
./test.sh 2>&1 | tee test_results_$(date +%Y%m%d_%H%M%S).log

# Check exit code
if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 ALL TESTS PASSED! Perfect!"
else
    echo ""
    echo "⚠️  Some tests failed - check the log above"
fi
