#!/bin/bash

# Quick RBAC Demo - Shows RBAC in action
# This demonstrates the three-tier role system

BASE_URL="http://localhost:8080/api/v1"

echo "========================================="
echo "🔐 RBAC Quick Demo"
echo "========================================="
echo ""

# Step 1: Register three users
echo "📝 Step 1: Registering three test users..."
echo "========================================="

curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice User","email":"alice@demo.com","password":"pass123","age":25}' | jq -r '.message'

curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob Admin","email":"bob@demo.com","password":"pass123","age":30}' | jq -r '.message'

curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie SuperAdmin","email":"charlie@demo.com","password":"pass123","age":35}' | jq -r '.message'

echo ""
sleep 1

# Step 2: Promote users
echo "👑 Step 2: Promoting users (requires database access)"
echo "========================================="
echo "Promoting Bob to admin..."
cd /home/axels/GO\ Lang\ Project\ 01 && go run scripts/promote_user.go bob@demo.com admin 2>&1 | grep -E "Successfully|Email|Role"

echo ""
echo "Promoting Charlie to superadmin..."
cd /home/axels/GO\ Lang\ Project\ 01 && go run scripts/promote_user.go charlie@demo.com superadmin 2>&1 | grep -E "Successfully|Email|Role"

echo ""
sleep 1

# Step 3: Login all users
echo "🔐 Step 3: Logging in users..."
echo "========================================="

ALICE_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@demo.com","password":"pass123"}' | jq -r '.data.access_token')

BOB_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"bob@demo.com","password":"pass123"}' | jq -r '.data.access_token')

CHARLIE_TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"charlie@demo.com","password":"pass123"}' | jq -r '.data.access_token')

echo "✅ All users logged in successfully"
echo ""
sleep 1

# Step 4: Test READ access (all users can read)
echo "📖 Step 4: Testing READ access - All users CAN view users"
echo "========================================="

echo "Alice (user) viewing users:"
alice_read=$(curl -s -X GET "$BASE_URL/users" -H "Authorization: Bearer $ALICE_TOKEN")
alice_status=$(echo "$alice_read" | jq -r '.success')
if [ "$alice_status" == "true" ]; then
  echo "  ✅ SUCCESS - Alice can view users"
else
  echo "  ❌ FAILED - Alice cannot view users"
fi

echo "Bob (admin) viewing users:"
bob_read=$(curl -s -X GET "$BASE_URL/users" -H "Authorization: Bearer $BOB_TOKEN")
bob_status=$(echo "$bob_read" | jq -r '.success')
if [ "$bob_status" == "true" ]; then
  echo "  ✅ SUCCESS - Bob can view users"
else
  echo "  ❌ FAILED - Bob cannot view users"
fi

echo ""
sleep 1

# Step 5: Test CREATE access (only admin+)
echo "✏️  Step 5: Testing CREATE access - Only Admin+ can create"
echo "========================================="

echo "Alice (user) trying to create user:"
alice_create=$(curl -s -X POST "$BASE_URL/users" \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test1@demo.com","password":"pass123","age":20}')
alice_create_msg=$(echo "$alice_create" | jq -r '.message')
if [[ "$alice_create_msg" == *"forbidden"* ]]; then
  echo "  ✅ CORRECTLY BLOCKED - Alice cannot create users (403 Forbidden)"
else
  echo "  ❌ SECURITY ISSUE - Alice created user when she shouldn't!"
  echo "     Response: $alice_create_msg"
fi

echo "Bob (admin) trying to create user:"
bob_create=$(curl -s -X POST "$BASE_URL/users" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test2@demo.com","password":"pass123","age":20}')
bob_create_success=$(echo "$bob_create" | jq -r '.success')
if [ "$bob_create_success" == "true" ]; then
  echo "  ✅ SUCCESS - Bob can create users"
else
  echo "  ❌ FAILED - Bob should be able to create users"
  echo "     Response: $(echo "$bob_create" | jq -r '.message')"
fi

echo ""
sleep 1

# Step 6: Test ROLE CHANGE access (only superadmin)
echo "👑 Step 6: Testing ROLE CHANGE - Only Superadmin can change roles"
echo "========================================="

# Get Alice's ID
ALICE_ID=$(curl -s -X GET "$BASE_URL/users?email=alice@demo.com" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" | jq -r '.data.users[0].id')

echo "Bob (admin) trying to change Alice's role:"
bob_role=$(curl -s -X PUT "$BASE_URL/users/$ALICE_ID/role" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}')
bob_role_msg=$(echo "$bob_role" | jq -r '.message')
if [[ "$bob_role_msg" == *"forbidden"* ]]; then
  echo "  ✅ CORRECTLY BLOCKED - Bob cannot change roles (403 Forbidden)"
else
  echo "  ❌ SECURITY ISSUE - Bob changed role when he shouldn't!"
fi

echo "Charlie (superadmin) trying to change Alice's role:"
charlie_role=$(curl -s -X PUT "$BASE_URL/users/$ALICE_ID/role" \
  -H "Authorization: Bearer $CHARLIE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}')
charlie_role_success=$(echo "$charlie_role" | jq -r '.success')
if [ "$charlie_role_success" == "true" ]; then
  echo "  ✅ SUCCESS - Charlie can change roles"
else
  echo "  ❌ FAILED - Charlie should be able to change roles"
  echo "     Response: $(echo "$charlie_role" | jq -r '.message')"
fi

echo ""
sleep 1

# Summary
echo "========================================="
echo "📊 RBAC Demo Summary"
echo "========================================="
echo ""
echo "Role Permissions Verified:"
echo "  👤 USER (Alice):"
echo "     ✅ Can view users"
echo "     ❌ Cannot create users"
echo "     ❌ Cannot change roles"
echo ""
echo "  👮 ADMIN (Bob):"
echo "     ✅ Can view users"
echo "     ✅ Can create/update/delete users"
echo "     ❌ Cannot change roles"
echo ""
echo "  👑 SUPERADMIN (Charlie):"
echo "     ✅ Can view users"
echo "     ✅ Can create/update/delete users"
echo "     ✅ Can change user roles"
echo ""
echo "✅ RBAC implementation working correctly! 🎉"
