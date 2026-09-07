package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var imageHTTPClient = &http.Client{Timeout: 180 * time.Second}

// Only the Go server sees OPENAI_API_KEY. Generation previews are not published automatically.
func imageHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Prompt string `json:"prompt"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	input.Prompt = strings.TrimSpace(input.Prompt)
	if len(input.Prompt) == 0 || len(input.Prompt) > 4000 {
		http.Error(w, "Enter a description up to 4000 bytes", 400)
		return
	}
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		http.Error(w, "AI generation is not configured yet. You can still upload your own media", 503)
		return
	}
	model := os.Getenv("OPENAI_IMAGE_MODEL")
	if model == "" {
		model = "gpt-image-2"
	}
	if model == "dall-e-3" {
		http.Error(w, "DALL·E 3 is retired. Configure a supported GPT Image model", 503)
		return
	}
	body, _ := json.Marshal(map[string]interface{}{"model": model, "prompt": input.Prompt, "n": 1, "size": "1024x1024", "quality": "low"})
	ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/images/generations", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Could not generate image", 500)
		return
	}
	request.Header.Set("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := imageHTTPClient.Do(request)
	if err != nil {
		http.Error(w, "Generation timed out or is unavailable. Please try again", 504)
		return
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		status := 502
		description := "Image generation failed. Please try another description"
		if response.StatusCode == 429 {
			status = 429
			description = "Image API credit or rate limit reached. Check API billing or try again later"
		}
		if response.StatusCode == 401 || response.StatusCode == 403 {
			status = 503
			description = "Image API access is not ready. Check the server API key and project verification"
		}
		http.Error(w, description, status)
		return
	}
	var result struct {
		Data []struct {
			Base64 string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 20<<20)).Decode(&result); err != nil || len(result.Data) == 0 || result.Data[0].Base64 == "" {
		http.Error(w, "No image was returned. Please retry", 502)
		return
	}
	writeJSON(w, 200, map[string]string{"image": "data:image/png;base64," + result.Data[0].Base64, "prompt": input.Prompt, "model": model})
}
