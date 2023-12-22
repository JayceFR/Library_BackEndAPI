package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginAccount struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *ApiHandler) HandleLoginAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	loginAccount := LoginAccount{}
	err := json.NewDecoder(r.Body).Decode(&loginAccount)
	if err != nil {
		fmt.Println(err.Error())
	}
	h := sha256.New()
	h.Write([]byte(loginAccount.Password))
	bs := h.Sum(nil)
	var check_account Account
	s.db.First(&check_account, "email = ? AND password = ?", loginAccount.Email, bs)
	return s.WriteJson(w, http.StatusOK, check_account)
}
