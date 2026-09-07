package handler

import (
	"context"
	"encoding/json"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/olivere/elastic/v7"
	"github.com/pborman/uuid"
	"net/http"
	"os"
	"regexp"
	"socialai/backend"
	"socialai/constants"
	"socialai/model"
	"socialai/service"
	"strings"
	"time"
)

var mySigningKey = []byte(os.Getenv("JWT_SECRET"))
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

type authKey struct{}
type identity struct {
	Username, SessionID string
	ExpiresAt           int64
}

func currentUser(r *http.Request) identity { return r.Context().Value(authKey{}).(identity) }
func decodeJSON(w http.ResponseWriter, r *http.Request, value interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(value); err != nil {
		http.Error(w, "Invalid request body", 400)
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
func signinHandler(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if !decodeJSON(w, r, &user) {
		return
	}
	user.Username = strings.TrimSpace(user.Username)
	if len(user.Password) > 72 || user.Username == "" || user.Password == "" {
		http.Error(w, "Invalid username or password", 401)
		return
	}
	success, err := service.CheckUser(user.Username, user.Password)
	if err != nil {
		http.Error(w, "Sign in is temporarily unavailable", 503)
		return
	}
	if !success {
		http.Error(w, "Invalid username or password", 401)
		return
	}
	sid := uuid.New()
	expires := time.Now().Add(24 * time.Hour).Unix()
	if err := backend.ESBackend.SaveToES(model.Session{Username: user.Username, ExpiresAt: expires}, constants.SESSION_INDEX, sid); err != nil {
		http.Error(w, "Could not create session", 503)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"username": user.Username, "sid": sid, "exp": expires, "iat": time.Now().Unix()})
	signed, err := token.SignedString(mySigningKey)
	if err != nil {
		http.Error(w, "Could not sign in", 500)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(signed))
}
func signupHandler(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if !decodeJSON(w, r, &user) {
		return
	}
	user.Username = strings.TrimSpace(user.Username)
	if !usernamePattern.MatchString(user.Username) || len(user.Password) < 8 || len(user.Password) > 72 {
		http.Error(w, "Use a 3–32 character username (letters, numbers, underscores) and an 8–72 byte password", 400)
		return
	}
	success, err := service.AddUser(&user)
	if err != nil {
		http.Error(w, "Registration is temporarily unavailable", 503)
		return
	}
	if !success {
		http.Error(w, "Username is already taken", 409)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func validSession(user identity) bool {
	if time.Now().Unix() >= user.ExpiresAt {
		return false
	}
	session, err := backend.ESBackend.ReadSession(user.SessionID)
	return err == nil && session.Username == user.Username && session.ExpiresAt > time.Now().Unix()
}
func requireAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.Fields(r.Header.Get("Authorization"))
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			http.Error(w, "Please log in", 401)
			return
		}
		token, err := (&jwt.Parser{ValidMethods: []string{"HS256"}}).Parse(header[1], func(t *jwt.Token) (interface{}, error) { return mySigningKey, nil })
		if err != nil || !token.Valid {
			http.Error(w, "Session expired. Please log in again", 401)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid session", 401)
			return
		}
		name, _ := claims["username"].(string)
		sid, _ := claims["sid"].(string)
		exp, _ := claims["exp"].(float64)
		user := identity{name, sid, int64(exp)}
		if name == "" || sid == "" || !validSession(user) {
			http.Error(w, "Session expired. Please log in again", 401)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), authKey{}, user)))
	})
}
func signoutHandler(w http.ResponseWriter, r *http.Request) {
	err := backend.ESBackend.DeleteFromES(constants.SESSION_INDEX, currentUser(r).SessionID)
	if err != nil && !elastic.IsNotFound(err) {
		http.Error(w, "Could not end session. Please retry", 503)
		return
	}
	w.WriteHeader(204)
}
