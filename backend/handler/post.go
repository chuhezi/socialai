package handler

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic/v7"
	"github.com/pborman/uuid"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"socialai/backend"
	"socialai/model"
	"socialai/service"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const maxMediaBytes = 10 << 20

var mediaTypes = map[string]string{".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".gif": "image/gif", ".webp": "image/webp", ".mp4": "video/mp4", ".webm": "video/webm", ".mov": "video/quicktime"}

func validMessage(message string) bool {
	n := utf8.RuneCountInString(message)
	return n > 0 && n <= 2000
}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMediaBytes+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, "Choose one image or video up to 10 MB", 400)
		return
	}
	defer r.MultipartForm.RemoveAll()
	caption := strings.TrimSpace(r.FormValue("message"))
	if !validMessage(caption) {
		http.Error(w, "Write a caption between 1 and 2000 characters", 400)
		return
	}
	if len(r.MultipartForm.File["media_file"]) != 1 {
		http.Error(w, "Choose one media file", 400)
		return
	}
	file, header, err := r.FormFile("media_file")
	if err != nil {
		http.Error(w, "Media file is not available", 400)
		return
	}
	defer file.Close()
	if header.Size <= 0 || header.Size > maxMediaBytes {
		http.Error(w, "Media must be between 1 byte and 10 MB", 413)
		return
	}
	expected, ok := mediaTypes[strings.ToLower(filepath.Ext(header.Filename))]
	if !ok {
		http.Error(w, "Use JPG, PNG, GIF, WebP, MP4, WebM or MOV", 400)
		return
	}
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		http.Error(w, "Cannot read media", 400)
		return
	}
	detected := http.DetectContentType(head[:n])
	// Go does not recognize every QuickTime brand; check its ISO base media header.
	quicktime := expected == "video/quicktime" && n >= 12 && string(head[4:8]) == "ftyp" && string(head[8:12]) == "qt  "
	if detected != expected && !quicktime {
		http.Error(w, "File contents do not match the selected image/video format", 400)
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Cannot read media", 400)
		return
	}
	post := model.Post{Id: uuid.New(), User: currentUser(r).Username, Message: caption, Type: strings.Split(expected, "/")[0], CreatedAt: time.Now().UnixMilli()}
	if err := service.SavePost(&post, file); err != nil {
		log.Printf("Save post %s failed: %v", post.Id, err)
		http.Error(w, "Could not save post. Please retry", 503)
		return
	}
	writeJSON(w, 201, post)
}
func postError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrPostPermissionDenied) {
		http.Error(w, "You can only change your own posts", 403)
	} else if elastic.IsNotFound(err) {
		http.Error(w, "Post no longer exists", 404)
	} else {
		http.Error(w, "Could not update post. Please retry", 503)
	}
}
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if err := service.DeletePost(mux.Vars(r)["id"], currentUser(r).Username); err != nil {
		postError(w, err)
		return
	}
	w.WriteHeader(204)
}
func editHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Message string `json:"message"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	input.Message = strings.TrimSpace(input.Message)
	if !validMessage(input.Message) {
		http.Error(w, "Write a caption between 1 and 2000 characters", 400)
		return
	}
	post, err := service.EditPost(r.Context(), mux.Vars(r)["id"], currentUser(r).Username, input.Message)
	if err != nil {
		postError(w, err)
		return
	}
	writeJSON(w, 200, post)
}
func likeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Liked bool `json:"liked"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	post, err := backend.ESBackend.SetLike(r.Context(), mux.Vars(r)["id"], currentUser(r).Username, input.Liked)
	if err != nil {
		postError(w, err)
		return
	}
	writeJSON(w, 200, post)
}

type searchOptions struct {
	user, keywords, mediaType string
	offset, limit             int
}

func parseSearch(r *http.Request) (searchOptions, error) {
	q := r.URL.Query()
	option := searchOptions{user: strings.TrimSpace(q.Get("user")), keywords: strings.TrimSpace(q.Get("keywords")), mediaType: q.Get("type"), limit: 24}
	if q.Get("mine") == "true" {
		option.user = currentUser(r).Username
	}
	if option.mediaType != "" && option.mediaType != "image" && option.mediaType != "video" {
		return option, fmt.Errorf("Invalid media type")
	}
	var err error
	if q.Has("limit") {
		option.limit, err = strconv.Atoi(q.Get("limit"))
		if err != nil {
			return option, fmt.Errorf("Invalid limit")
		}
	}
	if q.Has("offset") {
		option.offset, err = strconv.Atoi(q.Get("offset"))
		if err != nil {
			return option, fmt.Errorf("Invalid offset")
		}
	}
	if option.limit < 1 || option.limit > 100 || option.offset < 0 || option.offset+option.limit > 10000 || len(option.keywords) > 2000 || len(option.user) > 100 {
		return option, fmt.Errorf("Search is outside supported limits")
	}
	return option, nil
}
func searchHandler(w http.ResponseWriter, r *http.Request) {
	option, err := parseSearch(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	posts, err := service.SearchPosts(r.Context(), option.user, option.keywords, option.mediaType, option.offset, option.limit)
	if err != nil {
		http.Error(w, "Could not load posts. Please retry", 503)
		return
	}
	writeJSON(w, 200, posts)
}
