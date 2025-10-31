#!/bin/bash

# Test script for JWT Authentication endpoints
# Run the server first: ./bin/server

BASE_URL="http://localhost:8080/api/v1"
echo "🧪 Testing JWT Authentication Endpoints"
echo "========================================="

# Test 1: Register a new user
echo -e "\n1️⃣ Testing Registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "age": 30
  }')

echo "$REGISTER_RESPONSE" | jq '.'

# Extract access token
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.refresh_token')

if [ "$ACCESS_TOKEN" != "null" ]; then
    echo "✅ Registration successful! Got access token"
else
    echo "❌ Registration failed"
    exit 1
fi

# Test 2: Login with the same user
echo -e "\n2️⃣ Testing Login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }')

echo "$LOGIN_RESPONSE" | jq '.'

NEW_ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')

if [ "$NEW_ACCESS_TOKEN" != "null" ]; then
    echo "✅ Login successful! Got new access token"
else
    echo "❌ Login failed"
fi

# Test 3: Get Profile (Protected Route)
echo -e "\n3️⃣ Testing Get Profile (Protected)..."
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$PROFILE_RESPONSE" | jq '.'

USER_EMAIL=$(echo "$PROFILE_RESPONSE" | jq -r '.data.email')

if [ "$USER_EMAIL" == "john@example.com" ]; then
    echo "✅ Profile retrieved successfully!"
else
    echo "❌ Failed to get profile"
fi

# Test 4: Refresh Token
echo -e "\n4️⃣ Testing Token Refresh..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

echo "$REFRESH_RESPONSE" | jq '.'

NEW_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.data.access_token')

if [ "$NEW_TOKEN" != "null" ]; then
    echo "✅ Token refresh successful!"
else
    echo "❌ Token refresh failed"
fi

# Test 5: Test with invalid token
echo -e "\n5️⃣ Testing with Invalid Token..."
INVALID_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer invalid_token_here")

echo "$INVALID_RESPONSE" | jq '.'

if echo "$INVALID_RESPONSE" | jq -e '.success == false' > /dev/null; then
    echo "✅ Invalid token correctly rejected"
else
    echo "❌ Should reject invalid token"
fi

# Test 6: Test login with wrong password
echo -e "\n6️⃣ Testing Login with Wrong Password..."
WRONG_PASS_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "wrongpassword"
  }')

echo "$WRONG_PASS_RESPONSE" | jq '.'

if echo "$WRONG_PASS_RESPONSE" | jq -e '.success == false' > /dev/null; then
    echo "✅ Wrong password correctly rejected"
else
    echo "❌ Should reject wrong password"
fi

echo -e "\n========================================="
echo "✅ All authentication tests completed!"
echo "========================================="
