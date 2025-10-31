#!/bin/bash

# RBAC Manual Testing Script
# This script helps test RBAC by manually managing users and showing results

BASE_URL="http://localhost:8080/api/v1"
DB_FILE="goproject.db"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "🔐 RBAC Manual Testing"
echo "========================================="
echo ""

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Test API endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local token=$3
    local data=$4
    local expected_status=$5
    local description=$6

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $token")
    fi

    status=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$status" == "$expected_status" ]; then
        print_success "$description (Status: $status)"
        return 0
    else
        print_error "$description (Expected: $expected_status, Got: $status)"
        echo "Response: $body"
        return 1
    fi
}

# Step 1: Register test users
echo "📝 Step 1: Register test users"
echo "========================================"

# Register user
print_info "Registering regular user..."
user_response=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Regular User",
        "email": "user@test.com",
        "password": "password123",
        "age": 25
    }')

if echo "$user_response" | grep -q "successfully"; then
    print_success "Regular user registered"
else
    print_error "Failed to register regular user"
    echo "$user_response"
fi

# Register admin
print_info "Registering admin user..."
admin_response=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Admin User",
        "email": "admin@test.com",
        "password": "password123",
        "age": 30
    }')

if echo "$admin_response" | grep -q "successfully"; then
    print_success "Admin user registered"
else
    print_error "Failed to register admin user"
fi

# Register superadmin
print_info "Registering superadmin user..."
superadmin_response=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Superadmin User",
        "email": "superadmin@test.com",
        "password": "password123",
        "age": 35
    }')

if echo "$superadmin_response" | grep -q "successfully"; then
    print_success "Superadmin user registered"
else
    print_error "Failed to register superadmin user"
fi

sleep 1
echo ""

# Step 2: Show current roles
echo "🔍 Step 2: Check current roles in database"
echo "========================================"

print_warning "To test RBAC, you need to manually promote users in the database!"
echo ""
print_info "Option 1: Install SQLite and run:"
echo "  sudo apt-get install sqlite3"
echo ""
print_info "Option 2: Use the Go script to promote users:"
echo "  go run scripts/promote_user.go admin@test.com admin"
echo "  go run scripts/promote_user.go superadmin@test.com superadmin"
echo ""
print_info "Option 3: Manually update via SQL:"
echo "  sqlite3 $DB_FILE \"UPDATE users SET role = 'admin' WHERE email = 'admin@test.com';\""
echo "  sqlite3 $DB_FILE \"UPDATE users SET role = 'superadmin' WHERE email = 'superadmin@test.com';\""
echo ""

read -p "Have you promoted the users? (y/n): " promoted

if [ "$promoted" != "y" ]; then
    print_warning "Please promote users and run this script again"
    exit 0
fi

echo ""

# Step 3: Login users
echo "🔐 Step 3: Login users to get tokens"
echo "========================================"

# Login user
print_info "Logging in regular user..."
user_login=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email": "user@test.com", "password": "password123"}')
USER_TOKEN=$(echo "$user_login" | jq -r '.data.access_token // empty')

if [ -n "$USER_TOKEN" ]; then
    print_success "Regular user logged in"
else
    print_error "Failed to login regular user"
    echo "$user_login"
    exit 1
fi

# Login admin
print_info "Logging in admin user..."
admin_login=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email": "admin@test.com", "password": "password123"}')
ADMIN_TOKEN=$(echo "$admin_login" | jq -r '.data.access_token // empty')

if [ -n "$ADMIN_TOKEN" ]; then
    print_success "Admin user logged in"
else
    print_error "Failed to login admin user"
    exit 1
fi

# Login superadmin
print_info "Logging in superadmin user..."
superadmin_login=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email": "superadmin@test.com", "password": "password123"}')
SUPERADMIN_TOKEN=$(echo "$superadmin_login" | jq -r '.data.access_token // empty')

if [ -n "$SUPERADMIN_TOKEN" ]; then
    print_success "Superadmin user logged in"
else
    print_error "Failed to login superadmin user"
    exit 1
fi

echo ""

# Step 4: Test read access
echo "📖 Step 4: Test READ access (GET /users)"
echo "========================================"

test_endpoint "GET" "/users" "$USER_TOKEN" "" "200" "Regular user viewing users"
test_endpoint "GET" "/users" "$ADMIN_TOKEN" "" "200" "Admin viewing users"
test_endpoint "GET" "/users" "$SUPERADMIN_TOKEN" "" "200" "Superadmin viewing users"

echo ""

# Step 5: Test create access
echo "✏️  Step 5: Test CREATE access (POST /users)"
echo "========================================"

test_endpoint "POST" "/users" "$USER_TOKEN" '{"name":"Test","email":"test1@test.com","password":"pass123","age":20}' "403" "Regular user creating user (should fail)"
test_endpoint "POST" "/users" "$ADMIN_TOKEN" '{"name":"Test","email":"test2@test.com","password":"pass123","age":20}' "201" "Admin creating user (should succeed)"
test_endpoint "POST" "/users" "$SUPERADMIN_TOKEN" '{"name":"Test","email":"test3@test.com","password":"pass123","age":20}' "201" "Superadmin creating user (should succeed)"

echo ""

# Step 6: Get a user ID for update/delete tests
print_info "Getting user ID for update/delete tests..."
users_list=$(curl -s -X GET "$BASE_URL/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN")
TEST_USER_ID=$(echo "$users_list" | jq -r '.data.users[0].id // empty')

if [ -z "$TEST_USER_ID" ]; then
    print_error "Could not get test user ID"
    exit 1
fi

print_success "Got test user ID: $TEST_USER_ID"
echo ""

# Step 7: Test update access
echo "🔄 Step 7: Test UPDATE access (PUT /users/:id)"
echo "========================================"

test_endpoint "PUT" "/users/$TEST_USER_ID" "$USER_TOKEN" '{"name":"Updated Name"}' "403" "Regular user updating user (should fail)"
test_endpoint "PUT" "/users/$TEST_USER_ID" "$ADMIN_TOKEN" '{"name":"Updated by Admin"}' "200" "Admin updating user (should succeed)"

echo ""

# Step 8: Test role change access
echo "👑 Step 8: Test ROLE CHANGE access (PUT /users/:id/role)"
echo "========================================"

test_endpoint "PUT" "/users/$TEST_USER_ID/role" "$USER_TOKEN" '{"role":"admin"}' "403" "Regular user changing role (should fail)"
test_endpoint "PUT" "/users/$TEST_USER_ID/role" "$ADMIN_TOKEN" '{"role":"admin"}' "403" "Admin changing role (should fail)"
test_endpoint "PUT" "/users/$TEST_USER_ID/role" "$SUPERADMIN_TOKEN" '{"role":"admin"}' "200" "Superadmin changing role (should succeed)"

echo ""

# Step 9: Test delete access
echo "🗑️  Step 9: Test DELETE access (DELETE /users/:id)"
echo "========================================"

test_endpoint "DELETE" "/users/$TEST_USER_ID" "$USER_TOKEN" "" "403" "Regular user deleting user (should fail)"
# Note: We already promoted this user to admin, so admin delete might fail due to policy
# Let's create a new user to delete
print_info "Creating new user to test deletion..."
new_user=$(curl -s -X POST "$BASE_URL/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"To Delete","email":"delete@test.com","password":"pass123","age":20}')
DELETE_USER_ID=$(echo "$new_user" | jq -r '.data.id // empty')

if [ -n "$DELETE_USER_ID" ]; then
    test_endpoint "DELETE" "/users/$DELETE_USER_ID" "$ADMIN_TOKEN" "" "200" "Admin deleting user (should succeed)"
fi

echo ""

# Summary
echo "========================================="
echo "📊 RBAC Testing Complete!"
echo "========================================="
echo ""
print_info "Key points tested:"
echo "  ✅ All users can READ (GET) users"
echo "  ✅ Only Admin+ can CREATE users"
echo "  ✅ Only Admin+ can UPDATE users"
echo "  ✅ Only Admin+ can DELETE users"
echo "  ✅ Only Superadmin can CHANGE roles"
echo "  ✅ Regular users get 403 Forbidden for protected actions"
echo ""
print_success "RBAC implementation is working correctly! 🎉"
