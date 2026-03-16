package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/4ier/notion-cli/internal/mode"
	"github.com/4ier/notion-cli/internal/tui"
)

const (
	BaseURL        = "https://api.notion.com"
	NotionVersion  = "2026-03-11"
	DefaultTimeout = 30 * time.Second
)

type Client struct {
	token      string
	httpClient *http.Client
	debug      bool
	dryRun     bool
}

func New(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

func (c *Client) do(method, path string, body interface{}) ([]byte, error) {
	url := BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", NotionVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.dryRun {
		fmt.Printf("[dry-run] %s %s\n", method, url)
		if body != nil {
			data, _ := json.Marshal(body)
			fmt.Printf("[dry-run] Body: %s\n", string(data))
		}
		return nil, nil
	}

	if c.debug {
		fmt.Printf("→ %s %s\n", method, url)
	}

	stopSpinner := tui.StartSpinner(method + " " + path)
	resp, err := c.httpClient.Do(req)
	stopSpinner()
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.debug {
		fmt.Printf("← %d %s (%d bytes)\n", resp.StatusCode, resp.Status, len(respBody))
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			if mode.IsAgent() {
				return nil, fmt.Errorf("ERR:%s:%s", apiErr.Code, apiErr.Message)
			}
			hint := errorHint(apiErr.Code, apiErr.Message, path)
			if hint != "" {
				return nil, fmt.Errorf("%s: %s\n  → %s", apiErr.Code, apiErr.Message, hint)
			}
			return nil, fmt.Errorf("%s: %s", apiErr.Code, apiErr.Message)
		}
		if mode.IsAgent() {
			return nil, fmt.Errorf("ERR:http_%d:%s", resp.StatusCode, resp.Status)
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	return respBody, nil
}

func (c *Client) Get(path string) ([]byte, error) {
	return c.do("GET", path, nil)
}

// GetAll fetches all pages of a paginated GET endpoint by following next_cursor.
// Merges all results[] into a single list response.
func (c *Client) GetAll(path string) ([]byte, error) {
	return paginateGET(c, path)
}

// PostAll fetches all pages of a paginated POST endpoint (e.g. search, db query).
// body must be a map — start_cursor is injected automatically on each page.
func (c *Client) PostAll(path string, body map[string]interface{}) ([]byte, error) {
	return paginatePOST(c, path, body)
}

func paginateGET(c *Client, basePath string) ([]byte, error) {
	var allResults []interface{}
	cursor := ""
	pagesFetched := 0
	sep := "?"
	if strings.Contains(basePath, "?") {
		sep = "&"
	}

	for {
		path := basePath
		if cursor != "" {
			path = basePath + sep + "start_cursor=" + cursor
		}
		data, err := c.do("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var page map[string]interface{}
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, err
		}
		pagesFetched++

		if results, ok := page["results"].([]interface{}); ok {
			allResults = append(allResults, results...)
		}

		hasMore, _ := page["has_more"].(bool)
		if !hasMore {
			break
		}
		cursor, _ = page["next_cursor"].(string)
		if cursor == "" {
			break
		}
	}

	merged := map[string]interface{}{
		"object":        "list",
		"results":       allResults,
		"has_more":      false,
		"next_cursor":   nil,
		"pages_fetched": pagesFetched,
	}
	return json.Marshal(merged)
}

func paginatePOST(c *Client, path string, body map[string]interface{}) ([]byte, error) {
	var allResults []interface{}
	pagesFetched := 0

	// Work on a copy to avoid mutating the caller's map
	reqBody := make(map[string]interface{}, len(body))
	for k, v := range body {
		reqBody[k] = v
	}

	for {
		data, err := c.do("POST", path, reqBody)
		if err != nil {
			return nil, err
		}

		var page map[string]interface{}
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, err
		}
		pagesFetched++

		if results, ok := page["results"].([]interface{}); ok {
			allResults = append(allResults, results...)
		}

		hasMore, _ := page["has_more"].(bool)
		if !hasMore {
			break
		}
		cursor, _ := page["next_cursor"].(string)
		if cursor == "" {
			break
		}
		reqBody["start_cursor"] = cursor
	}

	merged := map[string]interface{}{
		"object":        "list",
		"results":       allResults,
		"has_more":      false,
		"next_cursor":   nil,
		"pages_fetched": pagesFetched,
	}
	return json.Marshal(merged)
}

func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	return c.do("POST", path, body)
}

func (c *Client) Patch(path string, body interface{}) ([]byte, error) {
	return c.do("PATCH", path, body)
}

func (c *Client) Delete(path string) ([]byte, error) {
	return c.do("DELETE", path, nil)
}

// GetMe returns the bot user info for the current token.
func (c *Client) GetMe() (map[string]interface{}, error) {
	data, err := c.Get("/v1/users/me")
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetUser retrieves a user by ID.
func (c *Client) GetUser(userID string) (map[string]interface{}, error) {
	data, err := c.Get("/v1/users/" + userID)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Search performs a search across the workspace.
func (c *Client) Search(query string, filter string, pageSize int, startCursor string) (map[string]interface{}, error) {
	body := map[string]interface{}{}
	if query != "" {
		body["query"] = query
	}
	if filter != "" {
		body["filter"] = map[string]interface{}{
			"value":    filter,
			"property": "object",
		}
	}
	if pageSize > 0 {
		body["page_size"] = pageSize
	}
	if startCursor != "" {
		body["start_cursor"] = startCursor
	}

	data, err := c.Post("/v1/search", body)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPage retrieves a page by ID.
func (c *Client) GetPage(pageID string) (map[string]interface{}, error) {
	data, err := c.Get("/v1/pages/" + pageID)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetBlock retrieves a single block by ID.
func (c *Client) GetBlock(blockID string) (map[string]interface{}, error) {
	data, err := c.Get("/v1/blocks/" + blockID)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetBlockChildren retrieves children of a block.
func (c *Client) GetBlockChildren(blockID string, pageSize int, startCursor string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v1/blocks/%s/children?page_size=%d", blockID, pageSize)
	if startCursor != "" {
		path += "&start_cursor=" + startCursor
	}
	data, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetDatabase retrieves a database by ID.
func (c *Client) GetDatabase(dbID string) (map[string]interface{}, error) {
	data, err := c.Get("/v1/databases/" + dbID)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryDatabase queries a database with filters and sorts.
func (c *Client) QueryDatabase(dbID string, body map[string]interface{}) (map[string]interface{}, error) {
	data, err := c.Post("/v1/databases/"+dbID+"/query", body)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetUsers lists all users.
func (c *Client) GetUsers(pageSize int, startCursor string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v1/users?page_size=%d", pageSize)
	if startCursor != "" {
		path += "&start_cursor=" + startCursor
	}
	data, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListComments lists comments on a block/page.
func (c *Client) ListComments(blockID string, pageSize int, startCursor string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v1/comments?block_id=%s&page_size=%d", blockID, pageSize)
	if startCursor != "" {
		path += "&start_cursor=" + startCursor
	}
	data, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// AddComment adds a comment to a page.
func (c *Client) AddComment(pageID, text string) ([]byte, error) {
	body := map[string]interface{}{
		"parent": map[string]interface{}{
			"page_id": pageID,
		},
		"rich_text": []map[string]interface{}{
			{"text": map[string]interface{}{"content": text}},
		},
	}
	return c.Post("/v1/comments", body)
}

// UploadFileContent sends file content to an existing file upload via multipart form.
func (c *Client) UploadFileContent(uploadID, fileName, contentType string, fileBytes []byte) ([]byte, error) {
	url := BaseURL + fmt.Sprintf("/v1/file_uploads/%s/send", uploadID)

	// Build multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	partHeader.Set("Content-Type", contentType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(fileBytes); err != nil {
		return nil, fmt.Errorf("write file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("finalize multipart body: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", NotionVersion)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if c.debug {
		fmt.Printf("→ POST %s (multipart, %d bytes)\n", url, body.Len())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.debug {
		fmt.Printf("← %d %s (%d bytes)\n", resp.StatusCode, resp.Status, len(respBody))
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// errorHint provides actionable suggestions for common API errors.
func errorHint(code, message, path string) string {
	switch code {
	case "object_not_found":
		if strings.Contains(path, "/v1/data_sources") {
			return "For data-sources commands, use a data_source_id (not a database_id).\n" +
				"  Run: notion search post-search --body '{\"filter\":{\"value\":\"data_source\",\"property\":\"object\"}}' to list them"
		}
		return "Check the ID is correct and the page/database is shared with your integration"
	case "unauthorized":
		return "Run 'notion auth login' to authenticate, or check your token"
	case "restricted_resource":
		return "Your integration doesn't have access. Share the page/database with your integration in Notion"
	case "rate_limited":
		return "Too many requests. Wait a moment and try again"
	case "validation_error":
		if strings.Contains(message, "is not a property") {
			return "Check property names with 'notion db view <id>' or 'notion page props <id>'"
		}
		if strings.Contains(message, "body failed validation") {
			return "Check your input format. Use --debug for request details"
		}
	case "conflict_error":
		return "The resource was modified by another process. Retry the operation"
	case "internal_server_error", "service_unavailable":
		return "Notion's servers are having issues. Try again in a few minutes"
	}
	return ""
}
