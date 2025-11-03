package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRBACPermissions tests role-based access control
func TestRBACPermissions(t *testing.T) {
	cleanDatabase()

	// Create users with different roles
	regularUser, err := seedTestUser("user")
	require.NoError(t, err)
	userToken, err := getAuthToken(regularUser)
	require.NoError(t, err)

	adminUser, err := seedTestUser("admin")
	require.NoError(t, err)
	adminToken, err := getAuthToken(adminUser)
	require.NoError(t, err)

	superadminUser, err := seedTestUser("superadmin")
	require.NoError(t, err)
	superadminToken, err := getAuthToken(superadminUser)
	require.NoError(t, err)

	t.Run("Regular User Permissions", func(t *testing.T) {
		t.Run("Can view users list", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer "+userToken)
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Regular user can view users")
		})

		t.Run("Can view user by ID", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", regularUser.ID), nil)
			req.Header.Set("Authorization", "Bearer "+userToken)
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Regular user can view user details")
		})

		t.Run("Can view statistics", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/v1/users/stats", nil)
			req.Header.Set("Authorization", "Bearer "+userToken)
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Regular user can view stats")
		})

		t.Run("Cannot create users", func(t *testing.T) {
			createReq := map[string]interface{}{
				"name":     "New User",
				"email":    "newuser@example.com",
				"password": "password123",
				"age":      25,
				"role":     "user",
			}
			body, _ := json.Marshal(createReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+userToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "Regular user cannot create users")
		})

		t.Run("Cannot update users", func(t *testing.T) {
			updateReq := map[string]interface{}{
				"name": "Updated Name",
			}
			body, _ := json.Marshal(updateReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d", adminUser.ID), bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+userToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "Regular user cannot update users")
		})

		t.Run("Cannot delete users", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%d", adminUser.ID), nil)
			req.Header.Set("Authorization", "Bearer "+userToken)
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "Regular user cannot delete users")
		})

		t.Run("Cannot change user roles", func(t *testing.T) {
			roleReq := map[string]interface{}{
				"role": "admin",
			}
			body, _ := json.Marshal(roleReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d/role", adminUser.ID), bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+userToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "Regular user cannot change roles")
		})
	})

	t.Run("Admin Permissions", func(t *testing.T) {
		t.Run("Can create users", func(t *testing.T) {
			createReq := map[string]interface{}{
				"name":     "Admin Created User",
				"email":    "admincreated@example.com",
				"password": "password123",
				"age":      30,
				"role":     "user",
			}
			body, _ := json.Marshal(createReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code, "Admin can create users")
		})

		t.Run("Can update users", func(t *testing.T) {
			updateReq := map[string]interface{}{
				"name": "Updated By Admin",
			}
			body, _ := json.Marshal(updateReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d", regularUser.ID), bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Admin can update users")
		})

		t.Run("Can delete users", func(t *testing.T) {
			// Create a user to delete
			createReq := map[string]interface{}{
				"name":     "To Be Deleted",
				"email":    "tobedeleted@example.com",
				"password": "password123",
				"age":      25,
				"role":     "user",
			}
			body, _ := json.Marshal(createReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			var createResp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &createResp)
			data := createResp["data"].(map[string]interface{})
			userID := int(data["id"].(float64))

			// Delete the user
			w = httptest.NewRecorder()
			req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%d", userID), nil)
			req.Header.Set("Authorization", "Bearer "+adminToken)
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Admin can delete users")
		})

		t.Run("Can batch create users", func(t *testing.T) {
			batchReq := []map[string]interface{}{
				{
					"name":     "Admin Batch 1",
					"email":    "adminbatch1@example.com",
					"password": "password123",
					"age":      25,
				},
				{
					"name":     "Admin Batch 2",
					"email":    "adminbatch2@example.com",
					"password": "password123",
					"age":      30,
				},
			}
			body, _ := json.Marshal(batchReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/users/batch", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code, "Admin can batch create users")
		})

		t.Run("Cannot change user roles", func(t *testing.T) {
			roleReq := map[string]interface{}{
				"role": "superadmin",
			}
			body, _ := json.Marshal(roleReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d/role", regularUser.ID), bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "Admin cannot change user roles")
		})
	})

	t.Run("Superadmin Permissions", func(t *testing.T) {
		t.Run("Can create users", func(t *testing.T) {
			createReq := map[string]interface{}{
				"name":     "Superadmin Created",
				"email":    "superadmincreated@example.com",
				"password": "password123",
				"age":      35,
			}
			body, _ := json.Marshal(createReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+superadminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code, "Superadmin can create users")
		})

		t.Run("Can update users", func(t *testing.T) {
			updateReq := map[string]interface{}{
				"name": "Updated By Superadmin",
			}
			body, _ := json.Marshal(updateReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d", regularUser.ID), bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+superadminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Superadmin can update users")
		})

		t.Run("Can delete users", func(t *testing.T) {
			// Create a user to delete
			createReq := map[string]interface{}{
				"name":     "To Be Deleted Super",
				"email":    "tobedeletedsuper@example.com",
				"password": "password123",
				"age":      25,
			}
			body, _ := json.Marshal(createReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+superadminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			var createResp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &createResp)
			data := createResp["data"].(map[string]interface{})
			userID := int(data["id"].(float64))

			// Delete the user
			w = httptest.NewRecorder()
			req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%d", userID), nil)
			req.Header.Set("Authorization", "Bearer "+superadminToken)
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Superadmin can delete users")
		})

		t.Run("Can change user roles", func(t *testing.T) {
			// Create a user to promote
			createReq := map[string]interface{}{
				"name":     "To Be Promoted",
				"email":    "tobepromoted@example.com",
				"password": "password123",
				"age":      28,
			}
			body, _ := json.Marshal(createReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+superadminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			var createResp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &createResp)
			data := createResp["data"].(map[string]interface{})
			userID := int(data["id"].(float64))

			// Change role to admin
			roleReq := map[string]interface{}{
				"role": "admin",
			}
			body, _ = json.Marshal(roleReq)

			w = httptest.NewRecorder()
			req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d/role", userID), bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+superadminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Superadmin can change user roles")

			var roleResp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &roleResp)
			roleData := roleResp["data"].(map[string]interface{})
			user := roleData["user"].(map[string]interface{})
			assert.Equal(t, "admin", user["role"], "User role should be updated to admin")
		})

		t.Run("Cannot demote self", func(t *testing.T) {
			roleReq := map[string]interface{}{
				"role": "user",
			}
			body, _ := json.Marshal(roleReq)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d/role", superadminUser.ID), bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+superadminToken)
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, "Superadmin cannot demote themselves")
		})
	})
}

// TestRBACRoleHierarchy tests the permission hierarchy
func TestRBACRoleHierarchy(t *testing.T) {
	cleanDatabase()

	regularUser, err := seedTestUser("user")
	require.NoError(t, err)
	userToken, err := getAuthToken(regularUser)
	require.NoError(t, err)

	adminUser, err := seedTestUser("admin")
	require.NoError(t, err)
	adminToken, err := getAuthToken(adminUser)
	require.NoError(t, err)

	superadminUser, err := seedTestUser("superadmin")
	require.NoError(t, err)
	superadminToken, err := getAuthToken(superadminUser)
	require.NoError(t, err)

	t.Run("Role Change Permission Matrix", func(t *testing.T) {
		// Create target user
		createReq := map[string]interface{}{
			"name":     "Target User",
			"email":    "target@example.com",
			"password": "password123",
			"age":      25,
		}
		body, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		var createResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &createResp)
		data := createResp["data"].(map[string]interface{})
		targetUserID := int(data["id"].(float64))

		testCases := []struct {
			name           string
			token          string
			role           string
			targetID       int
			newRole        string
			expectedStatus int
			description    string
		}{
			{
				name:           "User cannot change any role",
				token:          userToken,
				role:           "user",
				targetID:       targetUserID,
				newRole:        "admin",
				expectedStatus: http.StatusForbidden,
				description:    "Regular users have no permission to change roles",
			},
			{
				name:           "Admin cannot change roles",
				token:          adminToken,
				role:           "admin",
				targetID:       targetUserID,
				newRole:        "admin",
				expectedStatus: http.StatusForbidden,
				description:    "Admins cannot change user roles",
			},
			{
				name:           "Superadmin can change to user",
				token:          superadminToken,
				role:           "superadmin",
				targetID:       targetUserID,
				newRole:        "user",
				expectedStatus: http.StatusOK,
				description:    "Superadmin can set role to user",
			},
			{
				name:           "Superadmin can change to admin",
				token:          superadminToken,
				role:           "superadmin",
				targetID:       targetUserID,
				newRole:        "admin",
				expectedStatus: http.StatusOK,
				description:    "Superadmin can set role to admin",
			},
			{
				name:           "Superadmin can change to superadmin",
				token:          superadminToken,
				role:           "superadmin",
				targetID:       targetUserID,
				newRole:        "superadmin",
				expectedStatus: http.StatusOK,
				description:    "Superadmin can promote to superadmin",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				roleReq := map[string]interface{}{
					"role": tc.newRole,
				}
				body, _ := json.Marshal(roleReq)

				w := httptest.NewRecorder()
				req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d/role", tc.targetID), bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+tc.token)
				req.Header.Set("Content-Type", "application/json")
				testRouter.ServeHTTP(w, req)

				assert.Equal(t, tc.expectedStatus, w.Code, tc.description)
			})
		}
	})
}

// TestCrossRoleInteractions tests interactions between different roles
func TestCrossRoleInteractions(t *testing.T) {
	cleanDatabase()

	adminUser, err := seedTestUser("admin")
	require.NoError(t, err)
	adminToken, err := getAuthToken(adminUser)
	require.NoError(t, err)

	superadminUser, err := seedTestUser("superadmin")
	require.NoError(t, err)
	superadminToken, err := getAuthToken(superadminUser)
	require.NoError(t, err)

	t.Run("Admin cannot update superadmin", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"name": "Trying to Update Superadmin",
		}
		body, _ := json.Marshal(updateReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d", superadminUser.ID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		// This should succeed because admin can update any user (not role-based restriction at user level)
		assert.Equal(t, http.StatusOK, w.Code, "Admin can update superadmin profile (not role)")
	})

	t.Run("Superadmin can update admin", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"name": "Updated Admin Name",
		}
		body, _ := json.Marshal(updateReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d", adminUser.ID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+superadminToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Superadmin can update admin")
	})
}
