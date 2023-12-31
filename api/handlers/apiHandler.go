package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ApiHandler struct {
	db           gorm.DB
	conns        map[uuid.UUID]*websocket.Conn
	active_conns map[string]*websocket.Conn
}

type Account struct {
	ID          uuid.UUID `gorm:"primarykey" json:"id"`
	FirstName   string    `json:"first_name"`
	Email       string    `json:"email"`
	Password    []byte    `json:"password"`
	CommunityID uuid.UUID `gorm:"foreignkey:CommunityID" json:"community_id"`
}

type Community struct {
	ID            uuid.UUID `gorm:"primarykey" json:"id"`
	CommunityName string    `json:"community_name"`
	NoOfUsers     int       `json:"number_of_users"`
}

type Message struct {
	ID         uuid.UUID `gorm:"primarykey" json: "id"`
	Content    string    `json: "content"`
	SenderID   uuid.UUID `gorm:"foreignkey:SenderID" json: "sender_id"`
	ReceiverID uuid.UUID `gorm:"foreignkey:ReceiverID" json: "receiver_id"`
	SentAt     time.Time `json: "sent_at"`
	Seen       bool      `json: "seen"`
}

func New() *ApiHandler {
	//db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	dsn := "u217768772_Jayce:WHSJayce1@tcp(srv707.hstgr.io)/u217768772_Jayce?parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
	}
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&Community{})
	db.AutoMigrate(&Message{})
	return &ApiHandler{
		db:           *db,
		conns:        make(map[uuid.UUID]*websocket.Conn),
		active_conns: make(map[string]*websocket.Conn),
	}
}

func (s *ApiHandler) NewAccount(firstName string, email string, passowrd []byte) *Account {
	id := uuid.New()
	return &Account{
		ID:          id,
		FirstName:   firstName,
		Email:       email,
		Password:    passowrd,
		CommunityID: uuid.Nil,
	}
}

func (s *ApiHandler) NewCommunity(communityName string) *Community {
	id := uuid.New()
	return &Community{
		ID:            id,
		CommunityName: communityName,
	}
}

func (s *ApiHandler) WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func (s *ApiHandler) HandleAccount(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.handleGetAllAccount(ctx, w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *ApiHandler) HandleSpecificAccount(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.handleGetAccount(ctx, w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(ctx, w, r)
	}
	if r.Method == "POST" {
		fmt.Print("I am here")
		return s.handleUpdateCommId(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *ApiHandler) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "POST" {
		return s.HandleLoginAccount(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

// func (s *ApiHandler) HandleSpecificCommunity(w http.ResponseWriter, r *http.Request) error {
// 	ctx := context.Background()

// 	return fmt.Errorf("method not allowed %s", r.Method)
// }

func (s *ApiHandler) HandleComms(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.handleGetAllComms(ctx, w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateCommunity(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *ApiHandler) HandleMessages(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "POST" {
		return s.getMessageHandler(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *ApiHandler) HandleSpecificMessage(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.getChats(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}
