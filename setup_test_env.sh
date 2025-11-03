#!/bin/bash

################################################################################
# 🎯 SETUP: Prepare fresh environment for testing
################################################################################

echo "════════════════════════════════════════════════════════"
echo "  🚀 Setting up FRESH test environment"  
echo "════════════════════════════════════════════════════════"
echo ""

# 1. Kill all existing servers
echo "1️⃣  Stopping existing servers..."
pkill -9 -f "go run main.go" 2>/dev/null
pkill -9 -f "main.go" 2>/dev/null
lsof -ti:8080 | xargs kill -9 2>/dev/null
sleep 2
echo "   ✅ Servers stopped"
echo ""

# 2. Delete all databases
echo "2️⃣  Cleaning databases..."
rm -f goproject.db 2>/dev/null
rm -f cmd/api/goproject.db 2>/dev/null
echo "   ✅ Databases cleaned"
echo ""

# 3. Start fresh server
echo "3️⃣  Starting fresh server..."
cd cmd/api && nohup go run main.go > ../../server.log 2>&1 &
SERVER_PID=$!
cd ../..
sleep 5
echo "   ✅ Server started (PID: $SERVER_PID)"
echo ""

# 4. Wait for server to be ready
echo "4️⃣  Waiting for server to be ready..."
for i in {1..10}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "   ✅ Server is READY!"
        echo ""
        break
    fi
    echo "   ⏳ Attempt $i/10..."
    sleep 2
done

# 5. Verify server health
echo "5️⃣  Verifying server..."
HEALTH_STATUS=$(curl -s http://localhost:8080/health | jq -r '.status' 2>/dev/null)
if [ "$HEALTH_STATUS" == "healthy" ]; then
    echo "   ✅ Server is HEALTHY!"
else
    echo "   ❌ Server health check failed!"
    echo "   Check server.log for details"
    exit 1
fi
echo ""

echo "════════════════════════════════════════════════════════"
echo "  ✅ Environment ready! Run: ./test.sh"
echo "════════════════════════════════════════════════════════"
echo ""
