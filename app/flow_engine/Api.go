package flow_engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (e *FlowEngine) ExecuteAPICallStep(step FlowStep) error {
	// 1. Validasi parameter wajib
	if step.Params["url"] == nil || step.Params["method"] == nil {
		return fmt.Errorf("missing required params (url or method)")
	}

	// 2. Parse parameter
	url, err := e.renderTemplate(step.Params["url"].(string))
	if err != nil {
		return fmt.Errorf("error rendering url template: %v", err)
	}

	method := strings.ToUpper(step.Params["method"].(string))
	if !isValidHTTPMethod(method) {
		return fmt.Errorf("invalid HTTP method: %s", method)
	}

	// 3. Prepare headers
	headers := make(map[string]string)
	if step.Params["headers"] != nil {
		headersMap, ok := step.Params["headers"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("headers should be a key-value map")
		}

		for k, v := range headersMap {
			renderedValue, err := e.renderTemplate(fmt.Sprintf("%v", v))
			if err != nil {
				return fmt.Errorf("error rendering header %s: %v", k, err)
			}
			headers[k] = renderedValue
		}
	}

	// 4. Prepare body
	var body []byte
	if step.Params["body"] != nil {
		renderedBody, err := e.renderJSONTemplate(step.Params["body"])
		if err != nil {
			return fmt.Errorf("error rendering request body: %v", err)
		}
		body = renderedBody
	}

	// 5. Create HTTP request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 6. Execute request
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing API call: %v", err)
	}
	defer resp.Body.Close()

	// 7. Process response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	// 8. Store response if needed
	if storeKey, ok := step.Params["_store_result"].(string); ok {
		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("error parsing JSON response: %v", err)
		}

		// Add status code to result
		result["_status_code"] = resp.StatusCode
		result["_headers"] = resp.Header

		e.State[storeKey] = result
	}

	// 9. Check for error status codes
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned error status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Helper functions for API call
func isValidHTTPMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

func (e *FlowEngine) renderTemplate(templateStr string) (string, error) {
	// Simple template rendering with ${variable} substitution
	var result strings.Builder
	var inVar bool
	var varName strings.Builder

	for _, ch := range templateStr {
		if ch == '$' {
			inVar = true
			continue
		}

		if inVar {
			if ch == '{' {
				// Start of variable
				varName.Reset()
			} else if ch == '}' {
				// End of variable
				val, exists := e.State[varName.String()]
				if !exists {
					return "", fmt.Errorf("variable %s not found in state", varName.String())
				}
				result.WriteString(fmt.Sprintf("%v", val))
				inVar = false
			} else {
				// Accumulate variable name
				varName.WriteRune(ch)
			}
		} else {
			result.WriteRune(ch)
		}
	}

	if inVar {
		return "", fmt.Errorf("unclosed variable in template")
	}

	return result.String(), nil
}

func (e *FlowEngine) renderJSONTemplate(data interface{}) ([]byte, error) {
	// Convert to JSON first
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Render as template
	rendered, err := e.renderTemplate(string(jsonData))
	if err != nil {
		return nil, err
	}

	// Convert back to interface{} to handle nested templates
	var finalData interface{}
	if err := json.Unmarshal([]byte(rendered), &finalData); err != nil {
		return nil, err
	}

	return json.Marshal(finalData)
}
