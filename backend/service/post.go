package service

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"reflect"
	"time"

	"socialai/backend"
	"socialai/constants"
	"socialai/model"

	"cloud.google.com/go/storage"
	"github.com/olivere/elastic/v7"
)

var ErrPostPermissionDenied = errors.New("post belongs to another user")

func SearchPostsByUser(user string) ([]model.Post, error) {
	query := elastic.NewTermQuery("user", user)
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.POST_INDEX)
	if err != nil {
		return nil, err
	}
	return getPostFromSearchResult(searchResult), nil
}

func SearchPostsByKeywords(keywords string) ([]model.Post, error) {
	query := elastic.NewMatchQuery("message", keywords)
	query.Operator("AND")
	if keywords == "" {
		query.ZeroTermsQuery("all")
	}
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.POST_INDEX)
	if err != nil {
		return nil, err
	}
	return getPostFromSearchResult(searchResult), nil
}

func getPostFromSearchResult(searchResult *elastic.SearchResult) []model.Post {
	var ptype model.Post
	posts := []model.Post{}

	for _, item := range searchResult.Each(reflect.TypeOf(ptype)) {
		p := item.(model.Post)
		posts = append(posts, p)
	}
	return posts
}

func SavePost(post *model.Post, file multipart.File) error {
	medialink, err := backend.GCSBackend.SaveToGCS(file, post.Id)
	if err != nil {
		return err
	}
	post.Url = medialink

	if err := backend.ESBackend.SaveToES(post, constants.POST_INDEX, post.Id); err != nil {
		// Compensate a failed metadata write so uploads do not leave orphaned objects.
		if cleanupErr := backend.GCSBackend.DeleteFromGCS(post.Id); cleanupErr != nil {
			log.Printf("Upload cleanup required for %s: %v", post.Id, cleanupErr)
		}
		return err
	}
	return nil
}

func DeletePost(id string, username string) error {
	post, err := backend.ESBackend.ReadPostByID(id)
	if err != nil {
		return err
	}
	if post.User != username {
		return ErrPostPermissionDenied
	}

	if err := backend.GCSBackend.DeleteFromGCS(id); err != nil && !errors.Is(err, storage.ErrObjectNotExist) {
		return err
	}
	return backend.ESBackend.DeleteFromES(constants.POST_INDEX, id)
}

func SearchPosts(ctx context.Context, user, keywords, mediaType string, offset, limit int) ([]model.Post, error) {
	query := elastic.NewBoolQuery()
	if user != "" {
		query.Filter(elastic.NewTermQuery("user", user))
	}
	if keywords != "" {
		query.Must(elastic.NewMatchQuery("message", keywords).Operator("AND"))
	}
	// Legacy mapping has index:false but keyword doc values remain available.
	if mediaType != "" {
		query.Filter(elastic.NewScriptQuery(elastic.NewScript("doc['type'].size() > 0 && doc['type'].value == params.type").Param("type", mediaType)))
	}
	result, err := backend.ESBackend.SearchPosts(ctx, query, offset, limit)
	if err != nil {
		return nil, err
	}
	return getPostFromSearchResult(result), nil
}
func EditPost(ctx context.Context, id, username, message string) (*model.Post, error) {
	post, err := backend.ESBackend.ReadPostByID(id)
	if err != nil {
		return nil, err
	}
	if post.User != username {
		return nil, ErrPostPermissionDenied
	}
	return backend.ESBackend.UpdatePost(ctx, id, map[string]interface{}{"message": message, "updatedAt": time.Now().UnixMilli()})
}
