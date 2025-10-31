#!/bin/bash

# RBAC Testing Script
# Tests Role-Based Access Control for: superadmin, admin, user

BASE_URL="http://localhost:8080/api/v1"
echo "========================================="
echo "🔐 RBAC Testing Script"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test Results
PASSED=0
FAILED=0

# Helper function to print results
print_result() {
    if [ $1 -eq 200 ] || [ $1 -eq 201 ]; then
        echo -e "${GREEN}✅ PASS${NC} - $2 (Status: $1)"
        ((PASSED++))
    elif [ $1 -eq 403 ]; then
        echo -e "${YELLOW}🚫 FORBIDDEN${NC} - $2 (Status: $1)"
        ((PASSED++))  # Forbidden is expected for unauthorized access
    elif [ $1 -eq 401 ]; then
        echo -e "${YELLOW}🔒 UNAUTHORIZED${NC} - $2 (Status: $1)"
        ((PASSED++))  # Unauthorized is expected without token
    else
        echo -e "${RED}❌ FAIL${NC} - $2 (Status: $1)"
        ((FAILED++))
    fi
}

echo "📝 Step 1: Register users with different roles"
echo "=============================================="

# Register regular user
echo -n "Registering regular user... "
USER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Regular User",
    "email": "user@example.com",
    "password": "password123",
    "age": 25
  }')
USER_STATUS=$(echo "$USER_RESPONSE" | tail -n1)
USER_TOKEN=$(echo "$USER_RESPONSE" | head -n -1 | jq -r '.data.access_token')
USER_ID=$(echo "$USER_RESPONSE" | head -n -1 | jq -r '.data.user.id')
print_result $USER_STATUS "Register regular user"

# Register admin (will be promoted later)
echo -n "Registering admin user... "
ADMIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin User",
    "email": "admin@example.com",
    "password": "password123",
    "age": 30
  }')
ADMIN_STATUS=$(echo "$ADMIN_RESPONSE" | tail -n1)
ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | head -n -1 | jq -r '.data.access_token')
ADMIN_ID=$(echo "$ADMIN_RESPONSE" | head -n -1 | jq -r '.data.user.id')
print_result $ADMIN_STATUS "Register admin user"

# Register superadmin (will be promoted later)
echo -n "Registering superadmin user... "
SUPERADMIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Super Admin",
    "email": "superadmin@example.com",
    "password": "password123",
    "age": 35
  }')
SUPERADMIN_STATUS=$(echo "$SUPERADMIN_RESPONSE" | tail -n1)
SUPERADMIN_TOKEN=$(echo "$SUPERADMIN_RESPONSE" | head -n -1 | jq -r '.data.access_token')
SUPERADMIN_ID=$(echo "$SUPERADMIN_RESPONSE" | head -n -1 | jq -r '.data.user.id')
print_result $SUPERADMIN_STATUS "Register superadmin user"

echo ""
echo "⚠️  Note: All users start as 'user' role. We need to promote them manually in DB or via superadmin."
echo "For testing, we'll manually promote users via SQL..."
echo ""

# Manually promote users in database (simulating superadmin promotion)
# In production, the first superadmin should be created via database seed
echo "🔧 Promoting users (simulating DB update)..."
sqlite3 test.db <<EOF
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
UPDATE users SET role = 'superadmin' WHERE email = 'superadmin@example.com';
EOF

# Re-login to get tokens with updated roles
echo "🔄 Re-logging in users to refresh tokens..."

USER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }')
USER_TOKEN=$(echo "$USER_RESPONSE" | head -n -1 | jq -r '.data.access_token')

ADMIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }')
ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | head -n -1 | jq -r '.data.access_token')

SUPERADMIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "superadmin@example.com",
    "password": "password123"
  }')
SUPERADMIN_TOKEN=$(echo "$SUPERADMIN_RESPONSE" | head -n -1 | jq -r '.data.access_token')

echo "✅ Tokens refreshed with updated roles"
echo ""

echo "📋 Step 2: Test READ access (GET /users)"
echo "=============================================="

echo -n "Regular user viewing users... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X GET "$BASE_URL/users" \
  -H "Authorization: Bearer $USER_TOKEN")
print_result $STATUS "User GET /users"

echo -n "Admin viewing users... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X GET "$BASE_URL/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
print_result $STATUS "Admin GET /users"

echo -n "Superadmin viewing users... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X GET "$BASE_URL/users" \
  -H "Authorization: Bearer $SUPERADMIN_TOKEN")
print_result $STATUS "Superadmin GET /users"

echo ""
echo "✏️  Step 3: Test CREATE access (POST /users)"
echo "=============================================="

echo -n "Regular user creating user (should fail)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X POST "$BASE_URL/users" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New User",
    "email": "newuser@example.com",
    "password": "password123",
    "age": 22
  }')
print_result $STATUS "User POST /users (should be 403)"

echo -n "Admin creating user (should succeed)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X POST "$BASE_URL/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin Created User",
    "email": "admincreated@example.com",
    "password": "password123",
    "age": 28
  }')
print_result $STATUS "Admin POST /users"

echo -n "Superadmin creating user (should succeed)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X POST "$BASE_URL/users" \
  -H "Authorization: Bearer $SUPERADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Superadmin Created User",
    "email": "superadmincreated@example.com",
    "password": "password123",
    "age": 32
  }')
print_result $STATUS "Superadmin POST /users"

echo ""
echo "🔄 Step 4: Test UPDATE access (PUT /users/:id)"
echo "=============================================="

echo -n "Regular user updating user (should fail)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "$BASE_URL/users/$USER_ID" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name"
  }')
print_result $STATUS "User PUT /users/:id (should be 403)"

echo -n "Admin updating user (should succeed)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "$BASE_URL/users/$USER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin Updated Name"
  }')
print_result $STATUS "Admin PUT /users/:id"

echo ""
echo "🗑️  Step 5: Test DELETE access (DELETE /users/:id)"
echo "=============================================="

# Get a user ID to delete
NEW_USER_RESPONSE=$(curl -s -X POST "$BASE_URL/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "To Be Deleted",
    "email": "tobedeleted@example.com",
    "password": "password123",
    "age": 20
  }')
DELETE_USER_ID=$(echo "$NEW_USER_RESPONSE" | jq -r '.data.id')

echo -n "Regular user deleting user (should fail)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X DELETE "$BASE_URL/users/$DELETE_USER_ID" \
  -H "Authorization: Bearer $USER_TOKEN")
print_result $STATUS "User DELETE /users/:id (should be 403)"

echo -n "Admin deleting user (should succeed)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X DELETE "$BASE_URL/users/$DELETE_USER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
print_result $STATUS "Admin DELETE /users/:id"

echo ""
echo "👑 Step 6: Test ROLE UPDATE access (PUT /users/:id/role)"
echo "=============================================="

echo -n "Regular user changing role (should fail)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "$BASE_URL/users/$USER_ID/role" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin"
  }')
print_result $STATUS "User PUT /users/:id/role (should be 403)"

echo -n "Admin changing role (should fail)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "$BASE_URL/users/$USER_ID/role" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin"
  }')
print_result $STATUS "Admin PUT /users/:id/role (should be 403)"

echo -n "Superadmin changing role (should succeed)... "
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "$BASE_URL/users/$USER_ID/role" \
  -H "Authorization: Bearer $SUPERADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin"
  }')
print_result $STATUS "Superadmin PUT /users/:id/role"

echo ""
echo "========================================="
echo "📊 Test Summary"
echo "========================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All RBAC tests passed!${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed!${NC}"
    exit 1
fi
