package client

import (
	"context"
	"testing"
)

// TestWorkspaceIntegration tests workspace operations against a mock server
// Run this with: go run scripts/mock_server.go & go test -v ./internal/client -run TestWorkspaceIntegration
func TestWorkspaceIntegration(t *testing.T) {
	// Skip this test unless running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Configure client to use mock server
	config := &Config{
		PersonalAccessToken: "test-token",
		BaseURL:             "http://localhost:8080/api/v1",
		Region:              "us",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("CreateWorkspace", func(t *testing.T) {
		req := &CreateWorkspaceRequest{
			Name:                  "Test Workspace",
			NotificationEmails:    []string{"test@example.com"},
			ReturnWorkspaceAPIKey: true,
		}

		workspace, err := client.CreateWorkspace(ctx, req)
		if err != nil {
			t.Fatalf("CreateWorkspace failed: %v", err)
		}

		if workspace.Name != "Test Workspace" {
			t.Errorf("Expected name 'Test Workspace', got '%s'", workspace.Name)
		}

		if workspace.APIKey == "" {
			t.Error("Expected API key to be returned")
		}

		if len(workspace.NotificationEmails) != 1 || workspace.NotificationEmails[0] != "test@example.com" {
			t.Errorf("Expected notification emails ['test@example.com'], got %v", workspace.NotificationEmails)
		}

		// Store workspace ID for other tests
		workspaceID := workspace.ID

		t.Run("GetWorkspace", func(t *testing.T) {
			retrieved, err := client.GetWorkspace(ctx, workspaceID)
			if err != nil {
				t.Fatalf("GetWorkspace failed: %v", err)
			}

			if retrieved.ID != workspaceID {
				t.Errorf("Expected ID %d, got %d", workspaceID, retrieved.ID)
			}

			if retrieved.Name != "Test Workspace" {
				t.Errorf("Expected name 'Test Workspace', got '%s'", retrieved.Name)
			}
		})

		t.Run("UpdateWorkspace", func(t *testing.T) {
			updateReq := &UpdateWorkspaceRequest{
				Name:               "Updated Test Workspace",
				NotificationEmails: []string{"updated@example.com", "admin@example.com"},
			}

			updated, err := client.UpdateWorkspace(ctx, workspaceID, updateReq)
			if err != nil {
				t.Fatalf("UpdateWorkspace failed: %v", err)
			}

			if updated.Name != "Updated Test Workspace" {
				t.Errorf("Expected name 'Updated Test Workspace', got '%s'", updated.Name)
			}

			if len(updated.NotificationEmails) != 2 {
				t.Errorf("Expected 2 notification emails, got %d", len(updated.NotificationEmails))
			}
		})

		t.Run("ListWorkspaces", func(t *testing.T) {
			workspaces, pagination, err := client.ListWorkspaces(ctx, &ListOptions{
				Page:    1,
				PerPage: 25,
			})
			if err != nil {
				t.Fatalf("ListWorkspaces failed: %v", err)
			}

			if len(workspaces) == 0 {
				t.Error("Expected at least one workspace")
			}

			if pagination.TotalRecords == 0 {
				t.Error("Expected total records > 0")
			}

			// Find our workspace
			found := false
			for _, ws := range workspaces {
				if ws.ID == workspaceID {
					found = true
					if ws.Name != "Updated Test Workspace" {
						t.Errorf("Expected updated name in list")
					}
					break
				}
			}

			if !found {
				t.Error("Created workspace not found in list")
			}
		})

		t.Run("DeleteWorkspace", func(t *testing.T) {
			err := client.DeleteWorkspace(ctx, workspaceID)
			if err != nil {
				t.Fatalf("DeleteWorkspace failed: %v", err)
			}

			// Verify workspace is deleted
			_, err = client.GetWorkspace(ctx, workspaceID)
			if err == nil {
				t.Error("Expected error when getting deleted workspace")
			}

			// Check that it's a 404 error
			if apiErr, ok := err.(*APIError); ok {
				if apiErr.StatusCode != 404 {
					t.Errorf("Expected 404 error, got %d", apiErr.StatusCode)
				}
			} else {
				t.Errorf("Expected APIError, got %T", err)
			}
		})
	})
}
