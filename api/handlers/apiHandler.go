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

//Declaring the classes. 

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
	Request    bool      `json: "request"`
}

type Book struct {
	ID            uuid.UUID `gorm:"primarykey" json: "id"`
	Owner_id      uuid.UUID `gorm:"foreignkey:Owner_id json : "ownner_id"`
	ISBN          string    `json:isbn`
	Name          string    `json:"name"`
	Author        string    `json:"author"`
	Borrowed      bool      `json:"borrowed"`
	Borrowed_date time.Time `json:"date_borrowed"`
	Return_date   time.Time `json:"return_date"`
	Borrower_id   uuid.UUID `json:"borrower_id"`
}

type Images struct {
	ID        uuid.UUID `gorm:"primarykey" json:"id"`
	Object_id uuid.UUID `gorm:"foreignkey:Object_id" json:"object_id"`
	Type      string    `json:"type"`
	Data      []byte    `gorm:"type:longblob" json:"data"`
}

type Notifications struct {
	ID         uuid.UUID `gorm:"primarykey" json:"id"`
	Content    string    `json:"content"`
	Date       time.Time `json:"date"`
	ReceiverID uuid.UUID `json:"receiverid"`
}

//Constructor to create a new account object.
func New() *ApiHandler {
  //To host a database locally. 
	//db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	dsn := "u217768772_Jayce:WHSJayce1@tcp(srv707.hstgr.io)/u217768772_Jayce?parseTime=true"
  //Connecting to an external hosting database site. 
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
	}
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&Community{})
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&Book{})
	db.AutoMigrate(&Images{})
	db.AutoMigrate(&Notifications{})
	return &ApiHandler{
		db:           *db,
		conns:        make(map[uuid.UUID]*websocket.Conn),
		active_conns: make(map[string]*websocket.Conn),
	}
}

//Constructor to create a new acount object.
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

//Constructor to create a new community object. 
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

// Handling the /account endpoint. 
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

//Handling the /account/{id} endpoint. 
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

//Handling the /login endpoint. 
func (s *ApiHandler) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "POST" {
		return s.HandleLoginAccount(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

//Handling th /community endpoint 
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

//Handling the /community/{id} endpoint.
func (s *ApiHandler) HandleSpecificComm(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.handleGetCommunity(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

//Handling the /messages endpoint 
func (s *ApiHandler) HandleMessages(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "POST" {
		return s.getMessageHandler(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

//Handling the /messages{id} endpoint.
func (s *ApiHandler) HandleSpecificMessage(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.getChats(ctx, w, r)
	}
	if r.Method == "POST" {
		return s.handleUpdateMessage(ctx, w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteMessage(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

//Handling the /books endpoint. 
func (s *ApiHandler) HandleBooks(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.handle_get_books(ctx, w, r)
	}
	if r.Method == "POST" {
		return s.handle_post_book(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

//Handling the /book/{id} endpoint.
func (s *ApiHandler) HandleSpecificBook(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.handleGetSpecificBook(ctx, w, r)
	}
	if r.Method == "PUT" {
		return s.handleGetSpecificBookMessage(ctx, w, r)
	}
	if r.Method == "POST" {
		return s.handleUpdateBook(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

//Handling the /bookuser/{id} endpoint 
func (s *ApiHandler) HandleBookUser(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "GET" {
		return s.handleGetUserBook(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

//Handling the /images endpoint 
func (s *ApiHandler) HandleImages(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "POST" {
		return s.handleCreateImage(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

//Handling the /notifications endpoint. 
func (s *ApiHandler) HandleNotifications(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if r.Method == "DELETE" {
		return s.handeDeleteNotification(ctx, w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}
