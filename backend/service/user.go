package service

import (
	"crypto/subtle"
	"github.com/olivere/elastic/v7"
	"golang.org/x/crypto/bcrypt"
	"socialai/backend"
	"socialai/model"
	"strings"
)

func CheckUser(username, password string) (bool, error) {
	user, err := backend.ESBackend.ReadUser(username)
	if elastic.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if strings.HasPrefix(user.Password, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil, nil
	}
	// Read-only compatibility with classroom accounts shared by the older backend.
	// Migrating these passwords is a separate rollout, not a side effect of sign-in.
	return subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) == 1, nil
}
func AddUser(user *model.User) (bool, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	user.Password = string(hash)
	err = backend.ESBackend.CreateUser(user)
	if elastic.IsConflict(err) {
		return false, nil
	}
	return err == nil, err
}
