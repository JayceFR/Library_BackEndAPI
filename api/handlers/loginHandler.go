package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
)

//Used to handle the login endpoint.

type LoginAccount struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type return_account struct {
	User          Account          `json:"user"`
	Notifications []*Notifications `json:"notifications"`
}

func (s *ApiHandler) HandleLoginAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	loginAccount := LoginAccount{}
  //retrieve the data from the body of the request
	err := json.NewDecoder(r.Body).Decode(&loginAccount)
	if err != nil {
		fmt.Println(err.Error())
	}
  //hash the password 
	h := sha256.New()
	h.Write([]byte(loginAccount.Password))
	bs := h.Sum(nil)
  //check if the email address and the hashed password are found in the datbase. 
	var check_account Account
	s.db.First(&check_account, "email = ? AND password = ?", loginAccount.Email, bs)
  //fetch the unread notifications for the user
	notifications, err := s.handleGetNotifications(ctx, check_account.ID)
	if err != nil {
		return err
	}
	account := return_account{
		User:          check_account,
		Notifications: notifications,
	}
	return s.WriteJson(w, http.StatusOK, &account)
}
