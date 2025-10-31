#!/bin/bash

# Simple RBAC Test - Focused test without database promotion complexity

BASE_URL="http://localhost:8080/api/v1"

echo "========================================="
echo "🔐 Simple RBAC Test"
echo "========================================="
echo ""

# Step 1: Register one user
echo "📝 Registering test user..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@rbac.com","password":"pass123","age":25}' | jq -r '.message'

echo ""
sleep 1

# Step 2: Login
echo "🔐 Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@rbac.com","password":"pass123"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo "❌ Failed to get token"
  echo "$LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Login successful"
echo ""
sleep 1

# Step 3: Test GET (should succeed - all users can read)
echo "📖 Testing GET /users (should succeed)..."
GET_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/users" \
  -H "Authorization: Bearer $TOKEN")

GET_STATUS=$(echo "$GET_RESPONSE" | tail -n1)
GET_BODY=$(echo "$GET_RESPONSE" | sed '$d')

if [ "$GET_STATUS" == "200" ]; then
  echo "✅ GET request succeeded (Status: 200)"
  echo "   User can view users list"
else
  echo "❌ GET request failed (Status: $GET_STATUS)"
fi

echo ""
sleep 1

# Step 4: Test POST (should fail - only admin+)
echo "✏️  Testing POST /users (should fail with 403)..."
POST_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"New User","email":"new@test.com","password":"pass123","age":20}')

POST_STATUS=$(echo "$POST_RESPONSE" | tail -n1)
POST_BODY=$(echo "$POST_RESPONSE" | sed '$d')

if [ "$POST_STATUS" == "403" ]; then
  echo "✅ POST request correctly blocked (Status: 403 Forbidden)"
  echo "   Message: $(echo "$POST_BODY" | jq -r '.message')"
  echo "   🎉 RBAC IS WORKING CORRECTLY!"
elif [ "$POST_STATUS" == "201" ]; then
  echo "❌ SECURITY ISSUE: POST request succeeded when it should fail!"
  echo "   Regular user should NOT be able to create users"
  echo "   ⚠️  RBAC IS NOT WORKING!"
else
  echo "⚠️  Unexpected status: $POST_STATUS"
  echo "   Body: $POST_BODY"
fi

echo ""
sleep 1

# Step 5: Test role change (should fail - only superadmin)
echo "👑 Testing PUT /users/1/role (should fail with 403)..."
ROLE_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/users/1/role" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}')

ROLE_STATUS=$(echo "$ROLE_RESPONSE" | tail -n1)
ROLE_BODY=$(echo "$ROLE_RESPONSE" | sed '$d')

if [ "$ROLE_STATUS" == "403" ]; then
  echo "✅ Role change correctly blocked (Status: 403 Forbidden)"
  echo "   Message: $(echo "$ROLE_BODY" | jq -r '.message')"
else
  echo "❌ Unexpected status: $ROLE_STATUS"
  echo "   Body: $ROLE_BODY"
fi

echo ""
echo "========================================="
echo "📊 RBAC Test Summary"
echo "========================================="
echo ""

if [ "$GET_STATUS" == "200" ] && [ "$POST_STATUS" == "403" ] && [ "$ROLE_STATUS" == "403" ]; then
  echo "✅ ALL TESTS PASSED!"
  echo ""
  echo "Regular user (role=user) permissions verified:"
  echo "  ✅ CAN view users (GET)"
  echo "  ❌ CANNOT create users (POST) - correctly blocked"
  echo "  ❌ CANNOT change roles (PUT role) - correctly blocked"
  echo ""
  echo "🎉 RBAC implementation is working correctly!"
else
  echo "⚠️  SOME TESTS FAILED"
  echo ""
  echo "Results:"
  echo "  GET /users: $GET_STATUS $([ "$GET_STATUS" == "200" ] && echo "✅" || echo "❌")"
  echo "  POST /users: $POST_STATUS $([ "$POST_STATUS" == "403" ] && echo "✅" || echo "❌")"
  echo "  PUT /users/:id/role: $ROLE_STATUS $([ "$ROLE_STATUS" == "403" ] && echo "✅" || echo "❌")"
fi
