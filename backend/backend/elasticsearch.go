package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"socialai/constants"
	"socialai/model"

	"github.com/olivere/elastic/v7"
)

var (
	ESBackend *ElasticsearchBackend
)

type ElasticsearchBackend struct {
	client *elastic.Client
}

func InitElasticsearchBackend() {
	client, err := elastic.NewClient(
		elastic.SetURL(constants.ES_URL),
		elastic.SetSniff(false),
		elastic.SetHttpClient(&http.Client{Timeout: 15 * time.Second}),
		elastic.SetBasicAuth(constants.ES_USERNAME, constants.ES_PASSWORD))
	if err != nil {
		panic(err)
	}

	exists, err := client.IndexExists(constants.POST_INDEX).Do(context.Background())
	if err != nil {
		panic(err)
	}

	if !exists {
		mapping := `{
			"mappings": {
				"properties": {
					"id":       { "type": "keyword" },
					"user":     { "type": "keyword" },
					"message":  { "type": "text" },
					"url":      { "type": "keyword", "index": false },
					"type":     { "type": "keyword", "index": false }
				}
			}
		}`
		_, err := client.CreateIndex(constants.POST_INDEX).
			Body(mapping).
			Do(context.Background())
		if err != nil {
			panic(err)
		}
	}

	exists, err = client.IndexExists(constants.USER_INDEX).Do(context.Background())
	if err != nil {
		panic(err)
	}

	if !exists {
		mapping := `{
			"mappings": {
				"properties": {
					"username": {"type": "keyword"},
					"password": {"type": "keyword"},
					"age":      {"type": "long", "index": false},
					"gender":   {"type": "keyword", "index": false}
				}
			}
		}`
		_, err = client.CreateIndex(constants.USER_INDEX).
			Body(mapping).
			Do(context.Background())
		if err != nil {
			panic(err)
		}
	}

	// Additive mapping: existing posts remain valid and sort after dated posts.
	_, err = client.PutMapping().Index(constants.POST_INDEX).BodyString(`{"properties":{"createdAt":{"type":"long"},"updatedAt":{"type":"long"},"likes":{"type":"keyword","index":false}}}`).Do(context.Background())
	if err != nil {
		panic(err)
	}
	exists, err = client.IndexExists(constants.SESSION_INDEX).Do(context.Background())
	if err != nil {
		panic(err)
	}
	if !exists {
		_, err = client.CreateIndex(constants.SESSION_INDEX).BodyString(`{"mappings":{"properties":{"username":{"type":"keyword"},"expiresAt":{"type":"long"}}}}`).Do(context.Background())
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Indexes are ready.")

	ESBackend = &ElasticsearchBackend{client: client}
}

// Read data from Elasticsearch
func (backend *ElasticsearchBackend) ReadFromES(
	query elastic.Query,
	index string,
) (*elastic.SearchResult, error) {

	searchResult, err := backend.client.Search().
		Index(index).
		Query(query).
		Size(100).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	return searchResult, nil
}

func (backend *ElasticsearchBackend) SaveToES(i interface{}, index string, id string) error {
	_, err := backend.client.Index().
		Index(index).
		Id(id).
		BodyJson(i).
		Refresh("wait_for").
		Do(context.Background())
	return err
}

func (backend *ElasticsearchBackend) ReadPostByID(id string) (*model.Post, error) {
	result, err := backend.client.Get().Index(constants.POST_INDEX).Id(id).Do(context.Background())
	if err != nil {
		return nil, err
	}

	var post model.Post
	if err := json.Unmarshal(result.Source, &post); err != nil {
		return nil, err
	}
	return &post, nil
}

func (backend *ElasticsearchBackend) DeleteFromES(index string, id string) error {
	_, err := backend.client.Delete().Index(index).Id(id).Refresh("wait_for").Do(context.Background())
	return err
}

// Search stays bounded and deterministic, including legacy documents without dates.
func (backend *ElasticsearchBackend) SearchPosts(ctx context.Context, query elastic.Query, offset, limit int) (*elastic.SearchResult, error) {
	return backend.client.Search().Index(constants.POST_INDEX).Query(query).
		From(offset).Size(limit).SortBy(elastic.NewFieldSort("createdAt").Desc().Missing("_last").UnmappedType("long"), elastic.NewFieldSort("id").Asc()).Do(ctx)
}
func (backend *ElasticsearchBackend) UpdatePost(ctx context.Context, id string, fields map[string]interface{}) (*model.Post, error) {
	_, err := backend.client.Update().Index(constants.POST_INDEX).Id(id).Doc(fields).RetryOnConflict(5).Refresh("wait_for").Do(ctx)
	if err != nil {
		return nil, err
	}
	return backend.ReadPostByID(id)
}
func (backend *ElasticsearchBackend) SetLike(ctx context.Context, id, username string, liked bool) (*model.Post, error) {
	script := elastic.NewScript(`if (ctx._source.likes == null) { ctx._source.likes = []; } if (params.liked) { if (!ctx._source.likes.contains(params.user)) { ctx._source.likes.add(params.user); } } else { ctx._source.likes.removeIf(u -> u == params.user); }`).Params(map[string]interface{}{"user": username, "liked": liked})
	_, err := backend.client.Update().Index(constants.POST_INDEX).Id(id).Script(script).RetryOnConflict(5).Refresh("wait_for").Do(ctx)
	if err != nil {
		return nil, err
	}
	return backend.ReadPostByID(id)
}
func (backend *ElasticsearchBackend) ReadUser(username string) (*model.User, error) {
	result, err := backend.client.Get().Index(constants.USER_INDEX).Id(username).Do(context.Background())
	if err != nil {
		return nil, err
	}
	var user model.User
	err = json.Unmarshal(result.Source, &user)
	return &user, err
}
func (backend *ElasticsearchBackend) CreateUser(user *model.User) error {
	_, err := backend.client.Index().Index(constants.USER_INDEX).Id(user.Username).OpType("create").BodyJson(user).Refresh("wait_for").Do(context.Background())
	return err
}
func (backend *ElasticsearchBackend) ReadSession(id string) (*model.Session, error) {
	result, err := backend.client.Get().Index(constants.SESSION_INDEX).Id(id).Do(context.Background())
	if err != nil {
		return nil, err
	}
	var session model.Session
	err = json.Unmarshal(result.Source, &session)
	return &session, err
}
