package api

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *ApiHandler) GetAllAcounts(ctx context.Context, db gorm.DB) ([]*Account, error) {
	rows, err := s.db.WithContext(ctx).
		Select("*").
		Table("accounts").
		Rows()
	if err != nil {
		return []*Account{}, err
	}
	response := []*Account{}
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
}

func (s *ApiHandler) SearchAccount(ctx context.Context, db gorm.DB, query string) ([]*searchAccount, error) {
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
		Where("sender_id = ? ", sender_id).
		Or("sender_id = ? ", receiver_id).
		Where("receiver_id = ?", receiver_id).
		Or("receiver_id = ?", sender_id).
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
		)
		if err != nil {
			return []*Message{}, err
		}
		response = append(response, &message)
	}
	return response, nil
}
