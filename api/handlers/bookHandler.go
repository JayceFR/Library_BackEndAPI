package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type book_post struct {
	Owner_id string `gorm:"foreignkey:Owner_id json : "ownner_id"`
	ISBN     string `json:isbn`
	Name     string `json:"name"`
	Author   string `json:"author"`
}

func (s *ApiHandler) handle_post_book(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	book_to_post := book_post{}
	err := json.NewDecoder(r.Body).Decode(&book_to_post)
	if err != nil {
		return err
	}
	book := &Book{
		ID:          uuid.New(),
		Owner_id:    uuid.MustParse(book_to_post.Owner_id),
		ISBN:        book_to_post.ISBN,
		Name:        book_to_post.Name,
		Author:      book_to_post.Author,
		Borrowed:    false,
		Borrower_id: uuid.Nil,
	}

	s.db.Create(book)
	return s.WriteJson(w, http.StatusOK, book)
}
