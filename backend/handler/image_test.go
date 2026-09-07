package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type imageTransport func(*http.Request) (*http.Response, error)

func (transport imageTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return transport(r)
}

// Replace the transport: these tests never call OpenAI or spend API credits.
func TestImageAPIContract(t *testing.T) {
	original := imageHTTPClient
	t.Cleanup(func() { imageHTTPClient = original })
	t.Setenv("OPENAI_API_KEY", "local-test-key")
	t.Setenv("OPENAI_IMAGE_MODEL", "gpt-image-2")
	calls := 0
	upstreamStatus := 200
	upstreamBody := `{"data":[{"b64_json":"dGVzdA=="}]}`
	imageHTTPClient = &http.Client{Transport: imageTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.String() != "https://api.openai.com/v1/images/generations" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer local-test-key" {
			t.Fatal("Incorrect image API destination or authentication")
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["model"] != "gpt-image-2" || body["n"] != float64(1) || body["quality"] != "low" || body["size"] != "1024x1024" || body["prompt"] != "Mint coast" {
			t.Fatalf("Unexpected image request: %+v", body)
		}
		return &http.Response{StatusCode: upstreamStatus, Body: io.NopCloser(strings.NewReader(upstreamBody)), Header: make(http.Header)}, nil
	})}
	invoke := func(prompt string, want int) *httptest.ResponseRecorder {
		t.Helper()
		body, _ := json.Marshal(map[string]string{"prompt": prompt})
		w := httptest.NewRecorder()
		imageHandler(w, httptest.NewRequest("POST", "/generate", strings.NewReader(string(body))))
		if w.Code != want {
			t.Fatalf("Image response: got %d want %d: %s", w.Code, want, w.Body.String())
		}
		return w
	}
	invoke("   ", 400)
	invoke(strings.Repeat("x", 4001), 400)
	if calls != 0 {
		t.Fatal("Invalid prompts reached the paid API")
	}
	w := invoke(" Mint coast ", 200)
	if !strings.Contains(w.Body.String(), "data:image/png;base64,dGVzdA==") {
		t.Fatal("Missing image preview")
	}
	upstreamStatus = 429
	upstreamBody = `{"error":{"message":"private upstream diagnostic"}}`
	w = invoke("Mint coast", 429)
	if strings.Contains(w.Body.String(), "private upstream diagnostic") {
		t.Fatal("Exposed provider diagnostic")
	}
	upstreamStatus = 403
	invoke("Mint coast", 503)
	upstreamStatus = 200
	upstreamBody = `{"data":[]}`
	invoke("Mint coast", 502)
	t.Setenv("OPENAI_IMAGE_MODEL", "dall-e-3")
	before := calls
	invoke("Mint coast", 503)
	if calls != before {
		t.Fatal("Retired model reached the API")
	}
}
