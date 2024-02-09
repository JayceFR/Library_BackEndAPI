package api

// Handler for the /account endpoint

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *ApiHandler) handleGetAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"] //To extract something in the endpoint like '/account/{id}'
	var account *Account
	s.db.First(&account, "id = ?", id)
	return s.WriteJson(w, http.StatusOK, account)
}

type UpdateAccount struct {
	Community_ID string `json:"community_id"`
}

func (s *ApiHandler) handleUpdateCommId(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	updateAccount := UpdateAccount{}
	err := json.NewDecoder(r.Body).Decode(&updateAccount) // To extract the body of the request
	if err != nil {
		fmt.Printf(err.Error())
	}
	s.db.Model(&Account{}).Where("id = ?", id).Update("community_id", updateAccount.Community_ID)
	return s.WriteJson(w, http.StatusOK, updateAccount)

}

func (s *ApiHandler) handleGetAllAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	response, err := s.GetAllAcounts(ctx, s.db)
	if err != nil {
		fmt.Println(err.Error())
	}
	return s.WriteJson(w, http.StatusOK, response)
}

type CreateAccount struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

const null_uuid = "00000000-0000-0000-0000-000000000000"

func (s *ApiHandler) handleCreateAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	createAccount := CreateAccount{}
	err := json.NewDecoder(r.Body).Decode(&createAccount) //To extract the raw json body.
	if err != nil {
		fmt.Println(err.Error())
	}
	//Hashing the password using sha256.
	h := sha256.New()
	h.Write([]byte(createAccount.Password))
	bs := h.Sum(nil)
	var check_account Account
	// Checking if an account already exists with the same email in the database.
	s.db.First(&check_account, "email = ?", createAccount.Email)
	if check_account.ID.String() == null_uuid {
		fmt.Println("Email is not Found")
		//Create a new account and store it in the database.
		account := s.NewAccount(createAccount.FirstName, createAccount.Email, bs)
		s.db.Create(account)
		return s.WriteJson(w, http.StatusOK, createAccount)
	} else {
		return s.WriteJson(w, http.StatusBadRequest, "Found another user with the same email id")
	}
}

//Handle deleting the account from the database.
func (s *ApiHandler) handleDeleteAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	s.db.Delete(&Account{}, "id = ?", id)
	return s.WriteJson(w, http.StatusOK, "successfully deleted")
}
