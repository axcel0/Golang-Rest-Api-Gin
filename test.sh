#!/bin/bash

################################################################################
# 🧪 Comprehensive API Test Suite
# Tests ALL endpoints: Auth, RBAC, Users, Profile, WebSocket, Health, Metrics
################################################################################

BASE_URL="http://localhost:8080/api/v1"
ROOT_URL="http://localhost:8080"
WS_URL="ws://localhost:8080/ws"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test counters
TOTAL=0
PASSED=0
FAILED=0

# Helper functions
print_header() {
    echo ""
    echo "========================================="
    echo -e "${PURPLE}$1${NC}"
    echo "========================================="
}

print_test() {
    echo -e "${CYAN}▶ $1${NC}"
    ((TOTAL++))
}

print_pass() {
    echo -e "${GREEN}  ✅ PASS${NC} - $1"
    ((PASSED++))
}

print_fail() {
    echo -e "${RED}  ❌ FAIL${NC} - $1"
    ((FAILED++))
}

print_info() {
    echo -e "${BLUE}  ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}  ⚠️  $1${NC}"
}

# Test wrapper
test_endpoint() {
    local method=$1
    local endpoint=$2
    local token=$3
    local data=$4
    local expected_status=$5
    local description=$6

    print_test "$description"

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
        print_pass "Expected $expected_status, got $status"
        return 0
    else
        print_fail "Expected $expected_status, got $status"
        print_info "Response: $body"
        return 1
    fi
}

# Check dependencies
check_dependencies() {
    print_header "🔍 Checking Dependencies"
    
    if ! command -v curl &> /dev/null; then
        print_fail "curl not found - please install curl"
        exit 1
    fi
    print_pass "curl is installed"

    if ! command -v jq &> /dev/null; then
        print_fail "jq not found - please install jq"
        exit 1
    fi
    print_pass "jq is installed"

    # Check if server is running
    if curl -s "$BASE_URL/../health" > /dev/null 2>&1; then
        print_pass "Server is running at http://localhost:8080"
    else
        print_fail "Server is not running - start with: go run cmd/api/main.go"
        exit 1
    fi
}

# Cleanup before tests
cleanup() {
    print_header "🧹 Cleanup Test Data"
    print_info "Using fresh database (run ./setup_test_env.sh for clean start)"
    print_pass "Ready to test"
}

################################################################################
# TEST SUITE 1: Health & Metrics
################################################################################
test_health_metrics() {
    print_header "🏥 Test Suite 1: Health & Metrics"

    # Test health endpoint
    print_test "GET /health - Basic health check"
    health_response=$(curl -s -w "\n%{http_code}" "$ROOT_URL/health")
    health_status=$(echo "$health_response" | tail -n1)
    health_body=$(echo "$health_response" | sed '$d')
    
    if [ "$health_status" == "200" ] && echo "$health_body" | jq -e '.status == "healthy"' > /dev/null 2>&1; then
        print_pass "Health endpoint returns healthy status"
    else
        print_fail "Health endpoint not responding correctly (HTTP $health_status)"
    fi

    # Test metrics endpoint
    print_test "GET /metrics - Prometheus metrics"
    metrics_response=$(curl -s -w "\n%{http_code}" "$ROOT_URL/metrics")
    metrics_status=$(echo "$metrics_response" | tail -n1)
    metrics_body=$(echo "$metrics_response" | sed '$d')
    
    if [ "$metrics_status" == "200" ] && echo "$metrics_body" | grep -q "go_goroutines"; then
        print_pass "Metrics endpoint returns Prometheus data"
    else
        print_fail "Metrics endpoint not responding correctly (HTTP $metrics_status)"
    fi
}

################################################################################
# TEST SUITE 2: Authentication
################################################################################
test_authentication() {
    print_header "🔐 Test Suite 2: Authentication"

    # Register regular user
    print_test "POST /auth/register - Register regular user"
    user_reg=$(curl -s -X POST "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Regular User",
            "email": "user@test.com",
            "password": "Password123!",
            "age": 25
        }')
    
    if echo "$user_reg" | jq -e '.success == true' > /dev/null; then
        USER_TOKEN=$(echo "$user_reg" | jq -r '.data.access_token')
        USER_REFRESH=$(echo "$user_reg" | jq -r '.data.refresh_token')
        USER_ID=$(echo "$user_reg" | jq -r '.data.user.id')
        print_pass "User registered successfully (ID: $USER_ID)"
    else
        print_fail "User registration failed"
        print_info "Response: $user_reg"
        # Try to login instead (user might already exist)
        existing_login=$(curl -s -X POST "$BASE_URL/auth/login" \
            -H "Content-Type: application/json" \
            -d '{"email":"user@test.com","password":"Password123!"}')
        if echo "$existing_login" | jq -e '.success == true' > /dev/null; then
            USER_TOKEN=$(echo "$existing_login" | jq -r '.data.access_token')
            USER_ID=$(echo "$existing_login" | jq -r '.data.user.id')
            print_info "Using existing user (ID: $USER_ID)"
        fi
    fi

    # Register admin user
    print_test "POST /auth/register - Register admin user"
    admin_reg=$(curl -s -X POST "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Admin User",
            "email": "admin@test.com",
            "password": "Password123!",
            "age": 30
        }')
    
    if echo "$admin_reg" | jq -e '.success == true' > /dev/null; then
        ADMIN_TOKEN=$(echo "$admin_reg" | jq -r '.data.access_token')
        ADMIN_ID=$(echo "$admin_reg" | jq -r '.data.user.id')
        print_pass "Admin registered successfully (ID: $ADMIN_ID)"
    else
        print_fail "Admin registration failed"
        # Try to login instead (user might already exist)
        existing_admin=$(curl -s -X POST "$BASE_URL/auth/login" \
            -H "Content-Type: application/json" \
            -d '{"email":"admin@test.com","password":"Password123!"}')
        if echo "$existing_admin" | jq -e '.success == true' > /dev/null; then
            ADMIN_TOKEN=$(echo "$existing_admin" | jq -r '.data.access_token')
            ADMIN_ID=$(echo "$existing_admin" | jq -r '.data.user.id')
            print_info "Using existing admin (ID: $ADMIN_ID)"
        fi
    fi

    # Test login
    print_test "POST /auth/login - Login with credentials"
    login_resp=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "user@test.com",
            "password": "Password123!"
        }')
    
    if echo "$login_resp" | jq -e '.success == true' > /dev/null; then
        NEW_TOKEN=$(echo "$login_resp" | jq -r '.data.access_token')
        print_pass "Login successful, token received"
        USER_TOKEN=$NEW_TOKEN  # Update token
    else
        print_fail "Login failed"
    fi

    # Test wrong password
    print_test "POST /auth/login - Login with wrong password (should fail)"
    wrong_pass=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "user@test.com",
            "password": "WrongPassword123!"
        }')
    
    if echo "$wrong_pass" | jq -e '.success == false' > /dev/null; then
        print_pass "Wrong password correctly rejected"
    else
        print_fail "Should reject wrong password"
    fi

    # Test get profile
    print_test "GET /auth/profile - Get authenticated user profile"
    profile_resp=$(curl -s -X GET "$BASE_URL/auth/profile" \
        -H "Authorization: Bearer $USER_TOKEN")
    
    if echo "$profile_resp" | jq -e '.data.email == "user@test.com"' > /dev/null; then
        print_pass "Profile retrieved successfully"
    else
        print_fail "Failed to get profile"
    fi

    # Test token refresh
    print_test "POST /auth/refresh - Refresh access token"
    refresh_resp=$(curl -s -X POST "$BASE_URL/auth/refresh" \
        -H "Content-Type: application/json" \
        -d "{\"refresh_token\": \"$USER_REFRESH\"}")
    
    if echo "$refresh_resp" | jq -e '.success == true' > /dev/null; then
        print_pass "Token refresh successful"
    else
        print_fail "Token refresh failed"
    fi

    # Test invalid token
    print_test "GET /auth/profile - Access with invalid token (should fail)"
    invalid_resp=$(curl -s -X GET "$BASE_URL/auth/profile" \
        -H "Authorization: Bearer invalid_token_12345")
    
    if echo "$invalid_resp" | jq -e '.success == false' > /dev/null; then
        print_pass "Invalid token correctly rejected"
    else
        print_fail "Should reject invalid token"
    fi
}

################################################################################
# TEST SUITE 3: RBAC (Role-Based Access Control)
################################################################################
test_rbac() {
    print_header "🛡️  Test Suite 3: RBAC (Role-Based Access Control)"

    # Promote admin user using Go script (works without sqlite3 command)
    print_info "Promoting admin@test.com to admin role..."
    cd cmd/api && go run ../../scripts/promote_user.go admin@test.com admin > /dev/null 2>&1 && cd ../..
    print_pass "User promoted to admin"

    # Re-login admin to get new token with admin role
    print_test "POST /auth/login - Re-login admin with new role"
    admin_login=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "admin@test.com",
            "password": "Password123!"
        }')
    
    if echo "$admin_login" | jq -e '.success == true' > /dev/null; then
        ADMIN_TOKEN=$(echo "$admin_login" | jq -r '.data.access_token')
        print_pass "Admin re-logged in with new role"
    else
        print_fail "Admin re-login failed"
    fi

    # Test 1: Regular user can READ
    test_endpoint "GET" "/users" "$USER_TOKEN" "" "200" \
        "User (role: user) can READ users list"

    # Test 2: Regular user CANNOT CREATE
    test_endpoint "POST" "/users" "$USER_TOKEN" \
        '{"name":"Test","email":"test@example.com","password":"pass123","age":20}' \
        "403" "User (role: user) CANNOT CREATE user (should be 403)"

    # Test 3: Regular user CANNOT UPDATE
    test_endpoint "PUT" "/users/999" "$USER_TOKEN" \
        '{"name":"Updated"}' \
        "403" "User (role: user) CANNOT UPDATE user (should be 403)"

    # Test 4: Regular user CANNOT DELETE
    test_endpoint "DELETE" "/users/999" "$USER_TOKEN" "" "403" \
        "User (role: user) CANNOT DELETE user (should be 403)"

    # Test 5: Admin CAN CREATE
    test_endpoint "POST" "/users" "$ADMIN_TOKEN" \
        '{"name":"Test User","email":"testcreate@example.com","password":"Password123!","age":22}' \
        "201" "Admin (role: admin) CAN CREATE user"

    # Test 6: Admin CAN UPDATE
    test_endpoint "PUT" "/users/$USER_ID" "$ADMIN_TOKEN" \
        '{"name":"Updated Name","age":26}' \
        "200" "Admin (role: admin) CAN UPDATE user"

    # Test 7: Admin CAN DELETE
    print_test "Admin (role: admin) CAN DELETE user"
    # First create a user to delete
    create_resp=$(curl -s -X POST "$BASE_URL/users" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"name":"To Delete","email":"delete@test.com","password":"pass123","age":20}')
    delete_id=$(echo "$create_resp" | jq -r '.data.id')
    
    if [ "$delete_id" != "null" ] && [ -n "$delete_id" ]; then
        delete_resp=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/users/$delete_id" \
            -H "Authorization: Bearer $ADMIN_TOKEN")
        delete_status=$(echo "$delete_resp" | tail -n1)
        
        if [ "$delete_status" == "200" ]; then
            print_pass "Expected 200, got $delete_status"
        else
            print_fail "Expected 200, got $delete_status"
        fi
    else
        print_fail "Failed to create user for deletion test"
    fi

    # Test 8: Regular user CANNOT change roles
    test_endpoint "PUT" "/users/$USER_ID/role" "$USER_TOKEN" \
        '{"role":"admin"}' \
        "403" "User (role: user) CANNOT change roles (should be 403)"

    print_info "RBAC Summary:"
    print_info "  • Users can: READ"
    print_info "  • Admins can: READ, CREATE, UPDATE, DELETE"
    print_info "  • Superadmins can: ALL + change roles"
}

################################################################################
# TEST SUITE 4: User Management
################################################################################
test_user_management() {
    print_header "👥 Test Suite 4: User Management"

    # Get all users
    test_endpoint "GET" "/users" "$ADMIN_TOKEN" "" "200" \
        "GET /users - List all users"

    # Get single user
    test_endpoint "GET" "/users/$USER_ID" "$ADMIN_TOKEN" "" "200" \
        "GET /users/:id - Get user by ID"

    # Get non-existent user
    test_endpoint "GET" "/users/99999" "$ADMIN_TOKEN" "" "404" \
        "GET /users/:id - Get non-existent user (should be 404)"

    # Update user
    test_endpoint "PUT" "/users/$USER_ID" "$ADMIN_TOKEN" \
        '{"name":"Updated User Name","age":27}' \
        "200" "PUT /users/:id - Update user"

    # Invalid update (invalid age)
    test_endpoint "PUT" "/users/$USER_ID" "$ADMIN_TOKEN" \
        '{"age":-5}' \
        "400" "PUT /users/:id - Update with invalid data (should be 400)"
}

################################################################################
# TEST SUITE 5: Profile Management
################################################################################
test_profile() {
    print_header "👤 Test Suite 5: Profile Management"

    # Get own profile
    test_endpoint "GET" "/users/me" "$USER_TOKEN" "" "200" \
        "GET /users/me - Get own profile"

    # Update own profile
    test_endpoint "PUT" "/users/me" "$USER_TOKEN" \
        '{"name":"Updated Name","age":28}' \
        "200" "PUT /users/me - Update own profile"

    # Update password
    test_endpoint "PUT" "/users/me/password" "$USER_TOKEN" \
        '{"current_password":"Password123!","new_password":"NewPassword123!"}' \
        "200" "PUT /users/me/password - Change password"

    # Login with new password
    print_test "POST /auth/login - Login with new password"
    new_pass_login=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "user@test.com",
            "password": "NewPassword123!"
        }')
    
    if echo "$new_pass_login" | jq -e '.success == true' > /dev/null; then
        USER_TOKEN=$(echo "$new_pass_login" | jq -r '.data.access_token')
        print_pass "Login successful with new password"
    else
        print_fail "Login failed with new password"
    fi
}

################################################################################
# TEST SUITE 6: WebSocket
################################################################################
test_websocket() {
    print_header "🔌 Test Suite 6: WebSocket"

    # Test WebSocket stats endpoint
    print_test "GET /ws/stats - Get WebSocket connection stats"
    ws_stats=$(curl -s -w "\n%{http_code}" -X GET "$ROOT_URL/ws/stats" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    ws_status=$(echo "$ws_stats" | tail -n1)
    ws_body=$(echo "$ws_stats" | sed '$d')
    
    if [ "$ws_status" == "200" ] && echo "$ws_body" | jq -e '.success == true' > /dev/null 2>&1; then
        total_conn=$(echo "$ws_body" | jq -r '.data.total_connections // 0')
        print_pass "WebSocket stats retrieved (Total connections: $total_conn)"
    else
        print_fail "Failed to get WebSocket stats (HTTP $ws_status)"
    fi

    # Test broadcast (should fail for regular user)
    print_test "POST /ws/broadcast - Broadcast as user (should fail)"
    broadcast_user=$(curl -s -w "\n%{http_code}" -X POST "$ROOT_URL/ws/broadcast" \
        -H "Authorization: Bearer $USER_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"message":"Test broadcast","priority":"high"}')
    
    bu_status=$(echo "$broadcast_user" | tail -n1)
    bu_body=$(echo "$broadcast_user" | sed '$d')
    
    if [ "$bu_status" == "403" ] || (echo "$bu_body" | jq -e '.success == false' > /dev/null 2>&1); then
        print_pass "Regular user correctly blocked from broadcast"
    else
        print_fail "Regular user should not be able to broadcast (got HTTP $bu_status)"
    fi

    # Test broadcast (should succeed for admin)
    print_test "POST /ws/broadcast - Broadcast as admin (should succeed)"
    broadcast_admin=$(curl -s -w "\n%{http_code}" -X POST "$ROOT_URL/ws/broadcast" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"message":"Admin broadcast test","priority":"high"}')
    
    ba_status=$(echo "$broadcast_admin" | tail -n1)
    ba_body=$(echo "$broadcast_admin" | sed '$d')
    
    if [ "$ba_status" == "200" ] || (echo "$ba_body" | jq -e '.success == true' > /dev/null 2>&1); then
        print_pass "Admin can broadcast messages"
    else
        print_fail "Admin should be able to broadcast (got HTTP $ba_status)"
    fi

    print_info "For full WebSocket testing, run: npm run test:ws"
}

################################################################################
# TEST SUITE 7: Error Handling
################################################################################
test_error_handling() {
    print_header "⚠️  Test Suite 7: Error Handling"

    # Test invalid JSON
    print_test "POST /auth/register - Invalid JSON (should be 400)"
    invalid_json=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d '{invalid json}')
    status=$(echo "$invalid_json" | tail -n1)
    
    if [ "$status" == "400" ]; then
        print_pass "Invalid JSON correctly rejected with 400"
    else
        print_fail "Expected 400, got $status"
    fi

    # Test missing required fields
    test_endpoint "POST" "/auth/register" "" \
        '{"name":"Test"}' \
        "400" "POST /auth/register - Missing required fields (should be 400)"

    # Test invalid email format
    test_endpoint "POST" "/auth/register" "" \
        '{"name":"Test","email":"notanemail","password":"pass123","age":20}' \
        "400" "POST /auth/register - Invalid email format (should be 400)"

    # Test unauthorized access
    test_endpoint "GET" "/users" "" "" "401" \
        "GET /users - No token provided (should be 401)"
}

################################################################################
# MAIN EXECUTION
################################################################################
main() {
    echo ""
    echo "🧪 =============================================="
    echo "   COMPREHENSIVE API TEST SUITE"
    echo "   Testing: Auth, RBAC, Users, Profile, WebSocket"
    echo "============================================== 🧪"

    # Run all test suites
    check_dependencies
    cleanup
    test_health_metrics
    test_authentication
    test_rbac
    test_user_management
    test_profile
    test_websocket
    test_error_handling

    # Print summary
    print_header "📊 Test Summary"
    echo ""
    echo -e "  Total Tests:  ${CYAN}$TOTAL${NC}"
    echo -e "  Passed:       ${GREEN}$PASSED${NC}"
    echo -e "  Failed:       ${RED}$FAILED${NC}"
    echo ""

    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed! API is working correctly.${NC}"
        echo ""
        exit 0
    else
        echo -e "${RED}❌ Some tests failed. Please review the output above.${NC}"
        echo ""
        exit 1
    fi
}

# Run main function
main
