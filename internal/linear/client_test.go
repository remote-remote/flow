package linear

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssignedIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "test-key" {
			t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "test-key")
		}

		resp := map[string]any{
			"data": map[string]any{
				"viewer": map[string]any{
					"assignedIssues": map[string]any{
						"nodes": []map[string]any{
							{
								"id":         "id-1",
								"identifier": "ENG-1",
								"title":      "First issue",
								"url":        "https://linear.app/1",
								"state":      map[string]any{"name": "In Progress", "type": "started"},
							},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("test-key")
	c.httpClient = server.Client()
	// Override the URL by replacing the transport
	origDo := c.httpClient.Transport
	c.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		if origDo != nil {
			return origDo.RoundTrip(req)
		}
		return http.DefaultTransport.RoundTrip(req)
	})

	issues, err := c.AssignedIssues(context.Background())
	if err != nil {
		t.Fatalf("AssignedIssues: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("got %d issues, want 1", len(issues))
	}
	if issues[0].Identifier != "ENG-1" {
		t.Errorf("Identifier = %q, want %q", issues[0].Identifier, "ENG-1")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
