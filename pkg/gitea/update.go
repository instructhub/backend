package git

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"code.gitea.io/sdk/gitea"
)

type Operation string

const (
	OperationCreate Operation = "create"
	OperationDelete Operation = "delete"
	OperationUpdate Operation = "update"
)

// File represents a single file operation.
type File struct {
	Content   string    `json:"content"`
	Operation Operation `json:"operation"`
	Path      string    `json:"path"`
}

// ModifyRequest represents the JSON body structure for the API request.
type ModifyRequest struct {
	Author    gitea.Identity `json:"author"`
	Branch    string           `json:"branch"`
	Committer gitea.Identity `json:"committer"`
	Files     []File           `json:"files"`
	Message   string           `json:"message"`
	NewBranch string           `json:"new_branch"`
	Signoff   bool             `json:"signoff"`
}

// ModifyMultipleFiles sends a request to the Gitea API to modify multiple files.
func ModifyMultipleFiles(repoOwner, repoName string, request ModifyRequest) error {
	// Prepare the API URL
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/contents", os.Getenv("GITEA_URL"), repoOwner, repoName)

	// Marshal the request body into JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	httpRequest, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITEA_TOKEN")))

	// Send the HTTP request
	client := &http.Client{}
	response, err := client.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer response.Body.Close()

	// Check the response status
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("API request failed with status %d: %s", response.StatusCode, string(body))
	}

	// Success
	fmt.Println("Files modified successfully.")
	return nil
}
