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
	// rows, err := s.db.WithContext(ctx).
	// 	Distinct("a.first_name as first_name, CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END as id", user_id).
	// 	Table("messages, accounts as a").
	// 	Where("(sender_id = ? AND a.id = receiver_id)", user_id).
	// 	Or("(sender_id <> ? AND a.id = sender_id)", user_id).
	// 	Order("sent_at ASC").
	// 	Rows()
	subquery1 := s.db.Select("sender_id as user_id, max(sent_at) as latest_time").
		Table("messages").
		Where("sender_id <> ?", user_id).
		Group("sender_id")
	subquery2 := s.db.Select("receiver_id as user_id, max(sent_at) as latest_time").
		Table("messages").
		Where("receiver_id <> ?", user_id).
		Group("receiver_id")
	rows, err := s.db.WithContext(ctx).
		Distinct("a.first_name, a.id").
		Table("accounts as a").
		Joins("JOIN (? UNION ALL ?) as m ON (m.user_id = a.id)", subquery1, subquery2).
		Order("m.latest_time desc").
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
