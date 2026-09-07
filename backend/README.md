# Go API

The API keeps the classroom `handler / service / backend / model` structure and adds session revocation, caption editing, reactions, SSE, and authenticated image generation.

## Configuration

Copy `.env.example` to `.env.local`. Configure Elasticsearch, a GCS bucket, a random JWT signing secret of at least 32 bytes, and optionally an OpenAI API key for image generation. The local launcher reads this file as data, not shell code.

```sh
python3 scripts/run_local.py
```

The local API binds to `127.0.0.1:8080`. To use real Google Cloud Storage, configure Application Default Credentials with access to your own bucket (for example, `gcloud auth application-default login` for local development). The cloud runtime requires a service identity with the appropriate bucket permissions. Private Elasticsearch addresses must be reachable from the backend's network.

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/signup` | Create account; bcrypt password hashing |
| POST | `/signin` | Create session and issue a 24-hour JWT |
| POST | `/signout` | Revoke the current session |
| POST | `/upload` | Multipart `message` and `media_file`; returns the created post |
| GET | `/search` | `user`, `keywords`, `mine`, `type`, `limit`, `offset` |
| GET | `/events` | SSE using the same filters as search |
| PATCH | `/post/{id}` | Edit own caption with JSON `message` |
| PUT | `/post/{id}/like` | Set reaction with JSON `liked: true / false` |
| DELETE | `/post/{id}` | Delete own media and post |
| POST | `/generate` | Generate a preview using JSON `prompt` |
| GET | `/health` | Health/version marker |

All endpoints except signup, signin, and health require `Authorization: Bearer <token>`. The server validates the signing algorithm, expiration, and session record. Search is bounded to 100 results per request and `offset + limit <= 10000`. Writes wait for Elasticsearch search visibility.

## Tests

```sh
go test ./...
```

By default, integration tests are skipped. To run them, start dedicated local services:

```sh
docker run -d --name socialai-test-es -p 127.0.0.1:19200:9200 -e discovery.type=single-node -e xpack.security.enabled=false -e 'ES_JAVA_OPTS=-Xms256m -Xmx256m' docker.elastic.co/elasticsearch/elasticsearch:7.17.29
docker run -d --name socialai-test-gcs -p 127.0.0.1:14443:4443 fsouza/fake-gcs-server:1.52.3 -backend memory -scheme http -external-url http://127.0.0.1:14443 -public-host 127.0.0.1:14443
```

Wait for Elasticsearch to be ready, then run:

```sh
SOCIALAI_TEST_ES_URL=http://127.0.0.1:19200 STORAGE_EMULATOR_HOST=http://127.0.0.1:14443 go test -v ./...
```

These tests require the exact isolated loopback endpoints, generate their own fixtures, and do not call the real image-generation API. Remove only these test containers when finished.

## Deployment

`app.yaml` targets App Engine Flexible with Go 1.26 on Ubuntu 24, using 1–2 instances. Adapt its network settings to your own cloud project.

```sh
python3 scripts/prepare_deploy.py
```

This prepares a private temporary deployment directory; it does not deploy. Its generated `app.yaml` contains server configuration and must not be committed or shared. Deploy deliberately using your own project and inspect the new version before directing traffic to it. Running cloud instances incur charges even if they receive no default traffic.

At startup the API adds timestamp/reaction mappings and a session index. New sessions require a new login. Legacy classroom plaintext passwords remain read-compatible; migrate them separately before production use. Storage/search writes are not transactional, and the implementation has not been load-tested for production capacity.
