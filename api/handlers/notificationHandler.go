package api

// Handler for the /notification endpoint

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Constructor to create a new notification object
func (s *ApiHandler) newNotification(content string, receiver_id uuid.UUID) Notifications {
	notification := Notifications{
		ID:         uuid.New(),
		Content:    content,
		Date:       time.Now(),
		ReceiverID: receiver_id,
	}
	return notification
}

//handle storing the notification to the database. 
func (s *ApiHandler) handlePostNotifcation(notification Notifications) Notifications {
	s.db.Create(&notification)
	return notification
}

//handle fetching the notification from the database.
func (s *ApiHandler) handleGetNotifications(ctx context.Context, id uuid.UUID) ([]*Notifications, error) {
	rows, err := s.db.WithContext(ctx).
		Select("*").
		Table("notifications").
		Where("receiver_id = ?", id).
		Rows()
	if err != nil {
		return []*Notifications{}, err
	}
	notifications := []*Notifications{}
	for rows.Next() {
		notify := Notifications{}
		err := rows.Scan(
			&notify.ID,
			&notify.Content,
			&notify.Date,
			&notify.ReceiverID,
		)
		if err != nil {
			return []*Notifications{}, err
		}
		notifications = append(notifications, &notify) //store it in the slice. 
	}
	return notifications, nil
}

//handle deleting the notification from the database
func (s *ApiHandler) handeDeleteNotification(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	s.db.Delete(&Notifications{}, "id = ?", id)
	return s.WriteJson(w, http.StatusOK, "success")
}
