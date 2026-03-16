// file_upload_manual.go — NOT generated.
// Handles the multipart file upload flow which cannot be expressed via --body JSON.
// Flow: create-file → upload-file (multipart send) → complete-file-upload
package generated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

const notionAPIBase = "https://api.notion.com"
const notionVersion = "2026-03-11"

// newUploadFileCmd replaces the generated upload-file command with a real multipart implementation.
func newUploadFileCmd() *cobra.Command {
	var pageID string

	cmd := &cobra.Command{
		Use:   "upload-file <file-path>",
		Short: "Upload a file to Notion (create → send → complete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			token, err := GetToken()
			if err != nil {
				return err
			}

			// Read file
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
			fileName := filepath.Base(filePath)
			contentType := mime.TypeByExtension(filepath.Ext(filePath))
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			fmt.Printf("Uploading %s (%d bytes, %s)...\n", fileName, len(data), contentType)

			// Step 1: create file upload
			uploadID, err := createFileUpload(token, fileName, contentType)
			if err != nil {
				return fmt.Errorf("create file upload: %w", err)
			}
			fmt.Printf("  Created upload: %s\n", uploadID)

			// Step 2: send file content (multipart)
			if err := sendFileContent(token, uploadID, fileName, contentType, data); err != nil {
				return fmt.Errorf("send file content: %w", err)
			}
			fmt.Println("  Sent file content ✓")

			// Step 3: retrieve final status
			result, err := notionRequest("GET", fmt.Sprintf("/v1/file_uploads/%s", uploadID), token, nil)
			if err != nil {
				return fmt.Errorf("retrieve upload: %w", err)
			}
			fmt.Printf("  Status: %s ✓\n", result["status"])

			// Step 4: attach to page if requested
			if pageID != "" {
				_, err := notionRequest("PATCH", fmt.Sprintf("/v1/blocks/%s/children", pageID), token, map[string]interface{}{
					"children": []interface{}{
						map[string]interface{}{
							"object": "block",
							"type":   "image",
							"image": map[string]interface{}{
								"type": "file_upload",
								"file_upload": map[string]interface{}{
									"id": uploadID,
								},
							},
						},
					},
				})
				if err != nil {
					return fmt.Errorf("attach to page: %w", err)
				}
				fmt.Printf("  Attached to page %s ✓\n", pageID)
			}

			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(out))
			return nil
		},
	}

	cmd.Flags().StringVar(&pageID, "page-id", "", "Attach file to this page after upload")
	return cmd
}

func createFileUpload(token, name, contentType string) (string, error) {
	body := map[string]interface{}{
		"name":         name,
		"content_type": contentType,
	}
	resp, err := notionRequest("POST", "/v1/file_uploads", token, body)
	if err != nil {
		return "", err
	}
	id, ok := resp["id"].(string)
	if !ok {
		return "", fmt.Errorf("no id in response: %v", resp)
	}
	return id, nil
}

func sendFileContent(token, uploadID, fileName, contentType string, data []byte) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Create part with explicit Content-Type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
	h.Set("Content-Type", contentType)
	part, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	w.Close()

	url := notionAPIBase + fmt.Sprintf("/v1/file_uploads/%s/send", uploadID)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("send failed (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func completeFileUpload(token, uploadID string) (map[string]interface{}, error) {
	return notionRequest("POST", fmt.Sprintf("/v1/file_uploads/%s/complete", uploadID), token, nil)
}

func notionRequest(method, path, token string, body interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, notionAPIBase+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", notionVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var apiErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("%s: %s", apiErr.Code, apiErr.Message)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// addManualCommands injects manually-implemented commands into a generated group.
// Called from the generated register.go for every group — only acts on "file-uploads".
func addManualCommands(group *cobra.Command) {
	if group.Use == "file-uploads" {
		group.AddCommand(newUploadFileCmd())
	}
}
