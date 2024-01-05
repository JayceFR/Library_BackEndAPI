package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

type books_fetch struct {
	Book  Book   `json:"book"`
	Image Images `json:"image"`
}

func (s *ApiHandler) handle_get_books(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	books, err := s.GetBooks(ctx, s.db, "")
	fmt.Println(books[0].Book)
	if err != nil {
		return err
	}
	return s.WriteJson(w, http.StatusOK, books)
}

type specific_book_fetch struct {
	Book   Book      `json:"book"`
	Images []*Images `json:"images"`
}

func (s *ApiHandler) handleGetSpecificBook(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	book := Book{}
	s.db.First(&book, "id = ?", id)
	fmt.Println("Here is the book", book)
	images, err := s.handle_get_image(ctx, id)
	if err != nil {
		return err
	}
	return_data := specific_book_fetch{
		Book:   book,
		Images: images,
	}
	return s.WriteJson(w, http.StatusOK, return_data)
}

func (s *ApiHandler) handleGetSpecificBookMessage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	book := Book{}
	s.db.First(&book, "id = ?", id)
	return s.WriteJson(w, http.StatusOK, book)
}

type UpdateBook struct {
	State  string `json:"state"`
	From   string `json:"from"`
	To     string `json:"to"`
	UserId string `json:"userid"`
}

func (s *ApiHandler) handleUpdateBook(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	updateBook := UpdateBook{}
	err := json.NewDecoder(r.Body).Decode(&updateBook)
	if err != nil {
		return err
	}
	if updateBook.State == "return" {
		s.db.Model(&Book{}).Where("id = ?", id).Update("borrowed", 0)
	}
	if updateBook.State == "borrow" {
		book := Book{}
		s.db.First(&book, "id = ?", id)
		if !book.Borrowed {
			borrowed_date, err := time.Parse("2006-01-02 15:04:05.000", updateBook.From)
			if err != nil {
				return err
			}
			return_date, err := time.Parse("2006-01-02 15:04:05.000", updateBook.To)
			if err != nil {
				return err
			}
			s.db.Model(&Book{}).Where("id = ?", id).Update("borrowed", 1).Update("borrowed_date", borrowed_date).Update("return_date", return_date).Update("borrower_id", updateBook.UserId)
		} else {
			return fmt.Errorf("Already the book is borrwoed")
		}
	}
	return_book := Book{}
	s.db.First(&return_book, "id = ?", id)
	return s.WriteJson(w, http.StatusOK, return_book)
}
