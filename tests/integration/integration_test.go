package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestOrchestratorIntegration(t *testing.T) {
	expression := "(5+2) *(1-8) / (1-77)"
	requestBody, err := json.Marshal(map[string]string{"expression": expression})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post("http://localhost:8080/api/v1/calculate", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	t.Logf("Response body: %s", body)

	var result map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	id, exists := result["id"]
	if !exists {
		t.Fatalf("expected 'id' in response, got: %v", result)
	}

	idStr, ok := id.(string)
	if !ok {
		t.Fatalf("expected 'id' to be a string, but got: %T", id)
	}

	var found bool
	for i := 0; i < 10; i++ {
		resp, err = http.Get("http://localhost:8080/api/v1/expressions")
		if err != nil {
			t.Fatalf("failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var expressionsResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&expressionsResponse); err != nil {
			t.Fatalf("failed to decode result response: %v", err)
		}

		expressions, exists := expressionsResponse["expressions"]
		if !exists {
			t.Fatalf("expected 'expressions' in response, got: %v", expressionsResponse)
		}

		expressionsSlice, ok := expressions.([]interface{})
		if !ok {
			t.Fatalf("expected 'expressions' to be a slice, but got: %T", expressions)
		}

		for _, exp := range expressionsSlice {
			expMap, ok := exp.(map[string]interface{})
			if !ok {
				t.Fatalf("expected expression to be a map, but got: %T", exp)
			}

			if expMap["id"] == idStr {
				status, exists := expMap["status"]
				if !exists {
					t.Fatalf("expected 'status' in response, got: %v", expMap)
				}

				if status == "completed" {
					resultValue, exists := expMap["result"]
					if !exists {
						t.Fatalf("expected 'result' in response, got: %v", expMap)
					}

					resultFloat, ok := resultValue.(float64)
					if !ok {
						t.Fatalf("expected result to be a float64, but got: %T", resultValue)
					}

					if resultFloat != 0.644737 {
						t.Fatalf("expected result 0.644737, got %v", resultFloat)
					}

					found = true
					break
				}
			}
		}

		if found {
			break
		}

		time.Sleep(2 * time.Second)
	}

	if !found {
		t.Fatalf("task did not complete in time")
	}
}
