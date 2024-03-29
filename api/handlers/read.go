package api

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *ApiHandler) GetAllAcounts(ctx context.Context, db gorm.DB) ([]*Account, error) {
	//Get all the rows from the accounts table.
	rows, err := s.db.WithContext(ctx).
		Select("*").
		Table("accounts").
		Rows()
	//Check for any errors
	if err != nil {
		return []*Account{}, err
	}
	response := []*Account{}
	//Iterate though the rows and scan it into an Account object.
	for rows.Next() {
		account := Account{}
		err := rows.Scan(
			&account.ID,
			&account.FirstName,
			&account.Email,
			&account.Password,
			&account.CommunityID,
		)
		if err != nil {
			return []*Account{}, err
		}
		//Append the object to the list.
		response = append(response, &account)
	}
	return response, nil
}

func (s *ApiHandler) GetAllComms(ctx context.Context, db gorm.DB) ([]*Community, error) {
	//Fetch all the communities from the backend database.
	rows, err := s.db.WithContext(ctx).
		Select("*").
		Table("communities").
		Rows()
	if err != nil {
		return []*Community{}, err
	}
	response := []*Community{}
	//Turn the rows into a collection of objects
	for rows.Next() {
		comm := Community{}
		err := rows.Scan(
			&comm.ID,
			&comm.CommunityName,
			&comm.NoOfUsers,
		)
		if err != nil {
			return []*Community{}, err
		}
		response = append(response, &comm)
	}
	return response, nil
}

type searchAccount struct {
	ID        uuid.UUID `gorm:"primarykey" json:"id"`
	FirstName string    `json:"first_name"`
	Bubble    int64     `json:"bubble"`
	Active    bool      `json:"active"`
}

func (s *ApiHandler) SearchAccount(ctx context.Context, db gorm.DB, query string, id string) ([]*searchAccount, error) {
	rows, err := s.db.WithContext(ctx).
		Select("id, first_name").
		Table("accounts").
		Where("first_name like ?", "%"+query+"%").
		Rows()
	if err != nil {
		return []*searchAccount{}, err
	}
	response := []*searchAccount{}
	//turn it into json
	for rows.Next() {
		account := searchAccount{}
		err = rows.Scan(
			&account.ID,
			&account.FirstName,
		)
		_, ok := s.active_conns[account.ID.String()]
		if ok {
			account.Active = true
		} else {
			account.Active = false
		}
		s.db.Model(&Message{}).Where("sender_id = ?", account.ID).Where("receiver_id = ?", id).Where("seen = ?", 0).Count(&account.Bubble)
		if err != nil {
			return []*searchAccount{}, err
		}
		response = append(response, &account)
	}
	return response, nil
}

func (s *ApiHandler) GetMessages(ctx context.Context, db gorm.DB, sender_id string, receiver_id string) ([]*Message, error) {
	rows, err := s.db.WithContext(ctx).
		Select("*").
		Table("messages").
		Where("(sender_id = ? ", sender_id).
		Where("receiver_id = ?)", receiver_id).
		Or("(sender_id = ? ", receiver_id).
		Where("receiver_id = ?)", sender_id).
		Order("sent_at ASC").
		Rows()

	if err != nil {
		return []*Message{}, err
	}

	response := []*Message{}
	for rows.Next() {
		message := Message{}
		err = rows.Scan(
			&message.ID,
			&message.Content,
			&message.SenderID,
			&message.ReceiverID,
			&message.SentAt,
			&message.Seen,
			&message.Request,
		)
		if err != nil {
			return []*Message{}, err
		}
		response = append(response, &message)
	}
	return response, nil
}

// For the left pane in message system
func (s *ApiHandler) GetChatHistory(ctx context.Context, db gorm.DB, user_id string) ([]*searchAccount, error) {
	subquery1 := s.db.Select("sender_id as user_id, max(sent_at) as latest_time").
		Table("messages").
		Where("receiver_id = ?", user_id).
		Group("sender_id").
		Order("latest_time desc")
	subquery2 := s.db.Select("receiver_id as user_id, max(sent_at) as latest_time").
		Table("messages").
		Where("sender_id = ?", user_id).
		Group("receiver_id").
		Order("latest_time desc")
	rows, err := s.db.WithContext(ctx).
		Distinct("a.first_name, a.id").
		Table("accounts as a").
		Joins("JOIN ((?) UNION ALL (?) ORDER BY latest_time desc) as m ON m.user_id = a.id", subquery1, subquery2).
		Rows()
	if err != nil {
		return []*searchAccount{}, err
	}
	response := []*searchAccount{}
	for rows.Next() {
		account := searchAccount{}
		err = rows.Scan(
			&account.FirstName,
			&account.ID,
		)
		_, ok := s.active_conns[account.ID.String()]
		if ok {
			account.Active = true
		} else {
			account.Active = false
		}
		s.db.Model(&Message{}).
			Where("sender_id = ?", account.ID).
			Where("receiver_id = ?", user_id).
			Where("seen = ?", 0).
			Count(&account.Bubble)
		if err != nil {
			return []*searchAccount{}, err
		}
		response = append(response, &account)
	}
	return response, nil
}

//fetch the books for a respective user
func (s *ApiHandler) GetBooks(ctx context.Context, db gorm.DB, id string) ([]*books_fetch, error) {
	rows := &sql.Rows{}
	var err error
  //Check if an id of the user is provided. 
	if id == "" {
    //Fetch all the books 
		rows, err = s.db.WithContext(ctx).
			Select("*").
			Table("books").
			Rows()
	} else {
    //fetch a specific book.
		rows, err = s.db.WithContext(ctx).
			Select("*").
			Table("books").
			Where("owner_id = ?", id).
			Rows()
	}
	if err != nil {
		return []*books_fetch{}, err
	}
	books := []*books_fetch{}
	for rows.Next() {
		gbook := Book{}
    //store each individual book in the slice.
		err = rows.Scan(
			&gbook.ID,
			&gbook.Owner_id,
			&gbook.ISBN,
			&gbook.Name,
			&gbook.Author,
			&gbook.Borrowed,
			&gbook.Borrowed_date,
			&gbook.Return_date,
			&gbook.Borrower_id,
		)
		if err != nil {
			return []*books_fetch{}, err
		}
    //Fetch the profile image for the book
		gimage := Images{}
		s.db.First(&gimage, "object_id = ? and type = 'profile'", gbook.ID)
		curr_book := books_fetch{
			Book:  gbook,
			Image: gimage,
		}
		books = append(books, &curr_book)
	}
	return books, nil
}
