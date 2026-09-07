package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	jwt "github.com/form3tech-oss/jwt-go"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"socialai/backend"
	"socialai/constants"
	"socialai/model"
	"strings"
	"testing"
	"time"
)

// These tests only use a fresh, loopback-bound Elasticsearch and GCS emulator.
// Run with SOCIALAI_TEST_ES_URL=http://127.0.0.1:19200 and
// STORAGE_EMULATOR_HOST=http://127.0.0.1:14443. They never touch cloud content.
func TestLocalIntegration(t *testing.T) {
	endpoint := os.Getenv("SOCIALAI_TEST_ES_URL")
	if endpoint == "" {
		t.Skip("Set SOCIALAI_TEST_ES_URL to run isolated integration tests")
	}
	if endpoint != "http://127.0.0.1:19200" || os.Getenv("STORAGE_EMULATOR_HOST") != "http://127.0.0.1:14443" {
		t.Fatal("Integration tests require the isolated local endpoints")
	}
	constants.ES_URL = endpoint
	constants.ES_USERNAME = ""
	constants.ES_PASSWORD = ""
	constants.GCS_BUCKET = "socialai-local-test"
	mySigningKey = []byte("local-integration-only-secret-32-characters")
	backend.InitElasticsearchBackend()
	backend.InitGCSBackend()
	bucketReq, _ := http.NewRequest("POST", "http://127.0.0.1:14443/storage/v1/b", strings.NewReader(`{"name":"socialai-local-test"}`))
	bucketReq.Header.Set("Content-Type", "application/json")
	bucketResponse, err := http.DefaultClient.Do(bucketReq)
	if err != nil {
		t.Fatal(err)
	}
	bucketResponse.Body.Close()
	server := httptest.NewServer(InitRouter())
	defer server.Close()
	call := func(method, path, token string, data interface{}, want int) []byte {
		t.Helper()
		var body io.Reader
		if data != nil {
			encoded, _ := json.Marshal(data)
			body = bytes.NewReader(encoded)
		}
		req, _ := http.NewRequest(method, server.URL+path, body)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		raw, _ := io.ReadAll(res.Body)
		if res.StatusCode != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, res.StatusCode, want, raw)
		}
		return raw
	}
	suffix := fmt.Sprint(time.Now().UnixNano())
	owner := "owner_" + suffix
	other := "other_" + suffix
	ownerInput := map[string]string{"username": owner, "password": "local-test-password"}
	otherInput := map[string]string{"username": other, "password": "local-test-password"}
	call("GET", "/search", "", nil, 401)
	call("POST", "/signup", "", map[string]string{"username": "a", "password": "short"}, 400)
	call("POST", "/signup", "", ownerInput, 201)
	call("POST", "/signup", "", otherInput, 201)
	call("POST", "/signup", "", ownerInput, 409)
	stored, err := backend.ESBackend.ReadUser(owner)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stored.Password, "$2") {
		t.Fatal("Password must be hashed")
	}
	token := string(call("POST", "/signin", "", ownerInput, 200))
	otherToken := string(call("POST", "/signin", "", otherInput, 200))
	call("POST", "/signin", "", map[string]string{"username": owner, "password": "wrong"}, 401)
	legacyName := "legacy_" + suffix
	legacyUser := model.User{Username: legacyName, Password: "123"}
	if err := backend.ESBackend.SaveToES(legacyUser, constants.USER_INDEX, legacyName); err != nil {
		t.Fatal(err)
	}
	legacyToken := string(call("POST", "/signin", "", map[string]string{"username": legacyName, "password": "123"}, 200))
	storedLegacy, err := backend.ESBackend.ReadUser(legacyName)
	if err != nil || storedLegacy.Password != "123" {
		t.Fatal("Sign-in must not rewrite an account shared with the classroom backend")
	}
	call("POST", "/signout", legacyToken, nil, 204)
	upload := func(filename string, content []byte, want int) []byte {
		t.Helper()
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		writer.WriteField("message", "Teal seaside integration moment")
		part, _ := writer.CreateFormFile("media_file", filename)
		part.Write(content)
		writer.Close()
		req, _ := http.NewRequest("POST", server.URL+"/upload", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		data, _ := io.ReadAll(res.Body)
		if res.StatusCode != want {
			t.Fatalf("upload got %d want %d: %s", res.StatusCode, want, data)
		}
		return data
	}
	upload("fake.png", []byte("not an image"), 400)
	upload("file.txt", []byte("not media"), 400)
	pixel := image.NewRGBA(image.Rect(0, 0, 2, 2))
	pixel.Set(0, 0, color.RGBA{40, 124, 114, 255})
	var pngData bytes.Buffer
	png.Encode(&pngData, pixel)
	var post model.Post
	json.Unmarshal(upload("test.png", pngData.Bytes(), 201), &post)
	if post.Id == "" || post.User != owner || post.Type != "image" || post.CreatedAt == 0 || post.Url == "" {
		t.Fatalf("Incomplete created post: %+v", post)
	}
	media, err := http.Get(post.Url)
	if err != nil {
		t.Fatal(err)
	}
	mediaBytes, _ := io.ReadAll(media.Body)
	media.Body.Close()
	if media.StatusCode != 200 || !bytes.Equal(mediaBytes, pngData.Bytes()) {
		t.Fatal("Uploaded media cannot be read back")
	}
	// A new search must observe the acknowledged upload without a sleep.
	var posts []model.Post
	json.Unmarshal(call("GET", "/search?mine=true&type=image", token, nil, 200), &posts)
	if len(posts) != 1 || posts[0].Id != post.Id {
		t.Fatalf("Immediate own-post search: %+v", posts)
	}
	call("PATCH", "/post/"+post.Id, otherToken, map[string]string{"message": "Not mine"}, 403)
	call("DELETE", "/post/"+post.Id, otherToken, nil, 403)
	call("PATCH", "/post/"+post.Id, token, map[string]string{"message": "   "}, 400)
	call("PATCH", "/post/"+post.Id, token, map[string]string{"message": "Ocean mint edited caption"}, 200)
	call("GET", "/search?limit=101", token, nil, 400)
	call("GET", "/search?offset=-1", token, nil, 400)
	json.Unmarshal(call("GET", "/search?keywords=ocean+mint&type=image", token, nil, 200), &posts)
	if len(posts) == 0 || posts[0].Message != "Ocean mint edited caption" {
		t.Fatal("Edited caption was not searchable")
	}
	call("PUT", "/post/"+post.Id+"/like", otherToken, map[string]bool{"liked": true}, 200)
	raw := call("PUT", "/post/"+post.Id+"/like", otherToken, map[string]bool{"liked": true}, 200)
	var liked model.Post
	json.Unmarshal(raw, &liked)
	if len(liked.Likes) != 1 {
		t.Fatal("Repeated like created duplicates")
	}
	raw = call("PATCH", "/post/"+post.Id, token, map[string]string{"message": "Live updated caption"}, 200)
	json.Unmarshal(raw, &liked)
	if len(liked.Likes) != 1 {
		t.Fatal("Caption edit overwrote likes")
	}
	raw = call("PUT", "/post/"+post.Id+"/like", otherToken, map[string]bool{"liked": false}, 200)
	liked = model.Post{}
	json.Unmarshal(raw, &liked)
	if len(liked.Likes) != 0 {
		t.Fatal("Unlike failed")
	}
	// Verify media filtering and pagination past Elasticsearch's default ten hits.
	for i := 0; i < 13; i++ {
		item := model.Post{Id: fmt.Sprintf("%s-page-%02d", suffix, i), User: other, Message: "Pagination fixture", Type: "image", CreatedAt: int64(i + 1)}
		if err := backend.ESBackend.SaveToES(item, constants.POST_INDEX, item.Id); err != nil {
			t.Fatal(err)
		}
	}
	legacy := model.Post{Id: suffix + "-legacy", User: other, Message: "Undated classroom post", Type: "image"}
	if err := backend.ESBackend.SaveToES(legacy, constants.POST_INDEX, legacy.Id); err != nil {
		t.Fatal(err)
	}
	video := model.Post{Id: suffix + "-video", User: other, Message: "Video metadata fixture", Type: "video", CreatedAt: 100}
	if err := backend.ESBackend.SaveToES(video, constants.POST_INDEX, video.Id); err != nil {
		t.Fatal(err)
	}
	json.Unmarshal(call("GET", "/search?mine=true&type=image&limit=24", otherToken, nil, 200), &posts)
	if len(posts) != 14 || posts[0].CreatedAt != 13 || posts[13].Id != legacy.Id {
		t.Fatal("Media filter, newest-first ordering, or legacy pagination failed")
	}
	json.Unmarshal(call("GET", "/search?mine=true&type=image&limit=2&offset=12", otherToken, nil, 200), &posts)
	if len(posts) != 2 || posts[1].Id != legacy.Id {
		t.Fatal("Offset pagination failed")
	}
	json.Unmarshal(call("GET", "/search?mine=true&type=video", otherToken, nil, 200), &posts)
	if len(posts) != 1 || posts[0].Id != video.Id {
		t.Fatal("Video filter failed")
	}
	// Verify initial SSE payload and a later edit over a single connection.
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	streamReq, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/events?mine=true&type=image", nil)
	streamReq.Header.Set("Authorization", "Bearer "+token)
	stream, err := http.DefaultClient.Do(streamReq)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Body.Close()
	reader := bufio.NewReader(stream.Body)
	readData := func(want string) {
		t.Helper()
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("SSE waiting for %s: %v", want, err)
			}
			if strings.Contains(line, want) {
				return
			}
		}
	}
	readData("Live updated caption")
	call("PATCH", "/post/"+post.Id, token, map[string]string{"message": "Second window update"}, 200)
	readData("Second window update")
	cancel()
	// Missing API billing/key is an actionable response; no paid requests are made.
	t.Setenv("OPENAI_API_KEY", "")
	call("POST", "/generate", token, map[string]string{"prompt": "mint ocean"}, 503)
	call("DELETE", "/post/"+post.Id, token, nil, 204)
	json.Unmarshal(call("GET", "/search?mine=true", token, nil, 200), &posts)
	if len(posts) != 0 {
		t.Fatal("Deleted post remained in search")
	}
	call("POST", "/signout", token, nil, 204)
	call("GET", "/search", token, nil, 401)
	call("POST", "/signout", otherToken, nil, 204)
	t.Log("Verified real Elasticsearch writes/search, local media upload/delete, ownership, likes, SSE, revocation, and input validation")
}
func TestRejectInvalidJWT(t *testing.T) {
	mySigningKey = []byte("local-integration-only-secret-32-characters")
	handler := requireAuth(func(w http.ResponseWriter, r *http.Request) { t.Error("Invalid token reached protected handler") })
	for _, algorithm := range []jwt.SigningMethod{jwt.SigningMethodHS384, jwt.SigningMethodHS256} {
		expiry := time.Now().Add(-time.Hour).Unix()
		if algorithm == jwt.SigningMethodHS384 {
			expiry = time.Now().Add(time.Hour).Unix()
		}
		claims := jwt.MapClaims{"username": "owner", "sid": "unused", "exp": expiry}
		token, _ := jwt.NewWithClaims(algorithm, claims).SignedString(mySigningKey)
		req := httptest.NewRequest("GET", "/search", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 401 {
			t.Fatalf("Rejected token returned %d", w.Code)
		}
	}
}
