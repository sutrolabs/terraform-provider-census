package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Workspace struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	OrganizationID     int      `json:"organization_id"`
	CreatedAt          string   `json:"created_at"`
	NotificationEmails []string `json:"notification_emails"`
	APIKey             string   `json:"api_key,omitempty"`
}

type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

var workspaces = make(map[int]*Workspace)
var nextID = 1

func main() {
	http.HandleFunc("/api/v1/workspaces", handleWorkspaces)
	http.HandleFunc("/api/v1/workspaces/", handleWorkspaceByID)

	fmt.Println("🚀 Census API Mock Server starting on :8080")
	fmt.Println("📚 Available endpoints:")
	fmt.Println("   POST   /api/v1/workspaces       - Create workspace")
	fmt.Println("   GET    /api/v1/workspaces       - List workspaces")
	fmt.Println("   GET    /api/v1/workspaces/{id}  - Get workspace")
	fmt.Println("   PATCH  /api/v1/workspaces/{id}  - Update workspace")
	fmt.Println("   DELETE /api/v1/workspaces/{id}  - Delete workspace")
	fmt.Println()
	fmt.Println("🧪 Test with: export CENSUS_PERSONAL_ACCESS_TOKEN=test-token")
	fmt.Println("              go test ./... -v")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check authorization
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 401,
			"message": "Unauthorized",
		})
		return
	}

	switch r.Method {
	case http.MethodPost:
		handleCreateWorkspace(w, r)
	case http.MethodGet:
		handleListWorkspaces(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idStr := pathParts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetWorkspace(w, r, id)
	case http.MethodPatch:
		handleUpdateWorkspace(w, r, id)
	case http.MethodDelete:
		handleDeleteWorkspace(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name                  string   `json:"name"`
		NotificationEmails    []string `json:"notification_emails"`
		ReturnWorkspaceAPIKey bool     `json:"return_workspace_api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	workspace := &Workspace{
		ID:                 nextID,
		Name:               req.Name,
		OrganizationID:     1,
		CreatedAt:          time.Now().Format(time.RFC3339),
		NotificationEmails: req.NotificationEmails,
	}

	if req.ReturnWorkspaceAPIKey {
		workspace.APIKey = fmt.Sprintf("wsk_%d_test_api_key", nextID)
	}

	workspaces[nextID] = workspace
	nextID++

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{
		Status: "created",
		Data:   workspace,
	})

	fmt.Printf("✅ Created workspace: ID=%d, Name=%s\n", workspace.ID, workspace.Name)
}

func handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	var workspaceList []*Workspace
	for _, ws := range workspaces {
		workspaceList = append(workspaceList, ws)
	}

	response := struct {
		Status     string       `json:"status"`
		Data       []*Workspace `json:"data"`
		Pagination struct {
			TotalRecords int `json:"total_records"`
			PerPage      int `json:"per_page"`
			Page         int `json:"page"`
			LastPage     int `json:"last_page"`
		} `json:"pagination"`
	}{
		Status: "success",
		Data:   workspaceList,
	}

	response.Pagination.TotalRecords = len(workspaceList)
	response.Pagination.PerPage = 25
	response.Pagination.Page = 1
	response.Pagination.LastPage = 1

	json.NewEncoder(w).Encode(response)
	fmt.Printf("📋 Listed %d workspaces\n", len(workspaceList))
}

func handleGetWorkspace(w http.ResponseWriter, r *http.Request, id int) {
	workspace, exists := workspaces[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 404,
		})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Status: "success",
		Data:   workspace,
	})

	fmt.Printf("🔍 Retrieved workspace: ID=%d, Name=%s\n", workspace.ID, workspace.Name)
}

func handleUpdateWorkspace(w http.ResponseWriter, r *http.Request, id int) {
	workspace, exists := workspaces[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 404,
		})
		return
	}

	var req struct {
		Name               string   `json:"name"`
		NotificationEmails []string `json:"notification_emails"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	workspace.Name = req.Name
	workspace.NotificationEmails = req.NotificationEmails

	json.NewEncoder(w).Encode(Response{
		Status: "updated",
		Data:   workspace,
	})

	fmt.Printf("✏️ Updated workspace: ID=%d, Name=%s\n", workspace.ID, workspace.Name)
}

func handleDeleteWorkspace(w http.ResponseWriter, r *http.Request, id int) {
	_, exists := workspaces[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 404,
		})
		return
	}

	delete(workspaces, id)

	json.NewEncoder(w).Encode(Response{
		Status: "deleted",
	})

	fmt.Printf("🗑️ Deleted workspace: ID=%d\n", id)
}