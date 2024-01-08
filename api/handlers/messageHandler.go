package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

type Message_Posted struct {
	Content     string `json: "content"`
	Sender_ID   string `json: "sender_id"`
	Receiver_ID string `json: "receiver_id"`
	Request     bool   `json: "request"`
}

type GetMessage struct {
	SenderID   string `json: "sender_id"`
	ReceiverID string `json: "receiver_id"`
}

func (s *ApiHandler) getMessageHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	getMessage := GetMessage{}
	err := json.NewDecoder(r.Body).Decode(&getMessage)
	fmt.Println(getMessage.SenderID) // the current user who is using the application
	fmt.Println(getMessage.ReceiverID)
	if err != nil {
		return err
	}
	//make every message sent by the recepient and received by the user (SenderID) seen.
	s.db.Model(&Message{}).
		Where("sender_id = ? and receiver_id = ?", getMessage.ReceiverID, getMessage.SenderID).
		Updates(Message{Seen: true})
	response, erro := s.GetMessages(ctx, s.db, getMessage.SenderID, getMessage.ReceiverID)
	if erro != nil {
		return erro
	}
	return s.WriteJson(w, 200, response)
}

// Need to even get the number of unseen messages.
func (s *ApiHandler) getChats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	response, err := s.GetChatHistory(ctx, s.db, id)
	if err != nil {
		return err
	}
	return s.WriteJson(w, 200, response)
}

func (s *ApiHandler) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		fmt.Println("New incoming connection from client : ", ws.RemoteAddr())
		id := uuid.MustParse(r.URL.Query().Get("id"))
		s.conns[id] = ws
		s.readLoop(ws)
	})
	wsHandler.ServeHTTP(w, r)
}

func (s *ApiHandler) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Read error encountered : ", err)
			continue
		}
		msg := buf[:n]
		var post_message Message_Posted
		err = json.Unmarshal(msg, &post_message)
		if err != nil {
			panic(err)
		}
		fmt.Println("Sender id", post_message.Sender_ID)
		fmt.Println("Receiver id", post_message.Receiver_ID)
		fmt.Println("Content", post_message.Content)

		fmt.Println(string(msg))
		// ws.Write([]byte("thank you for the msg!!! "))
		if !post_message.Request {
			message := &Message{
				ID:         uuid.New(),
				Content:    post_message.Content,
				SenderID:   uuid.MustParse(post_message.Sender_ID),
				ReceiverID: uuid.MustParse(post_message.Receiver_ID),
				SentAt:     time.Now(),
				Seen:       false,
				Request:    false,
			}
			s.broadcast_message(message, ws)
		} else {
			message := &Message{
				ID:         uuid.New(),
				Content:    post_message.Content, //holds the uuid of the book
				SenderID:   uuid.MustParse(post_message.Sender_ID),
				ReceiverID: uuid.MustParse(post_message.Receiver_ID),
				SentAt:     time.Now(),
				Seen:       false,
				Request:    true,
			}
			s.broadcast_message(message, ws)
		}
	}
}

func (s *ApiHandler) handleGetUserBook(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	books, err := s.GetBooks(ctx, s.db, id)
	if err != nil {
		return err
	}
	return s.WriteJson(w, http.StatusOK, books)
}

type UpdateMessage struct {
	Content string `json:"content"`
}

func (s *ApiHandler) handleUpdateMessage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	update_message := UpdateMessage{}
	err := json.NewDecoder(r.Body).Decode(&update_message)
	if err != nil {
		return err
	}
	s.db.Model(&Message{}).Where("id = ?", id).Update("content", update_message.Content)
	return_data := Message{}
	s.db.First(&return_data, "id = ?", id)
	return s.WriteJson(w, http.StatusOK, return_data)
}

func (s *ApiHandler) handleDeleteMessage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	s.db.Delete(&Message{}, "id = ?", id)
	return s.WriteJson(w, http.StatusOK, "success")
}

func (s *ApiHandler) broadcast_message(message *Message, sender *websocket.Conn) {
	recepient, ok := s.conns[message.ReceiverID]
	marshal_message, erro := json.Marshal(message)
	if erro != nil {
		fmt.Println("Error while marshalling the message structure : ", erro)
	}
	if !ok {
		fmt.Println("Recepient not found ")

	} else {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(marshal_message); err != nil {
				fmt.Println("Write error : ", err)
			}
		}(recepient)
	}
	go func(ws *websocket.Conn) {
		if _, err := ws.Write(marshal_message); err != nil {
			fmt.Println("Write error : ", err)
		}
	}(sender)
	//Write the message to the database in either case
	s.db.Create(message)
}

// active endpoint
func (s *ApiHandler) ActiveWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		fmt.Println("New incoming active connection from client : ", ws.RemoteAddr())
		s.activeLoop(ws)
	})
	wsHandler.ServeHTTP(w, r)
}

type broad_id struct {
	ID string `json:"id"`
}

type input_broad struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type return_message struct {
	Type         string        `json:"type"`
	Conns        []*broad_id   `json:"conns"`
	Notification Notifications `json:"notification"`
}

func (s *ApiHandler) activeLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Read error encountered : ", err)
			continue
		}
		msg := buf[:n]
		var broad input_broad
		err = json.Unmarshal(msg, &broad)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(msg))
		if broad.Type == "add" {
			s.active_conns[broad.ID] = ws
			s.broadcast_to_all(ws)
		}
		if broad.Type == "notify" {
			//Add it to the db
			notification := s.newNotification(broad.Content, uuid.MustParse(broad.ID))
			recepient, ok := s.active_conns[broad.ID]
			if ok {
				//Send it to the user in real time.
				data := return_message{
					Type:         "notify",
					Notification: notification,
				}
				marshal_message, erro := json.Marshal(data)
				if erro != nil {
					fmt.Println("Error while marshalling the message structure : ", erro)
				}
				go func(ws *websocket.Conn) {
					if _, err := ws.Write(marshal_message); err != nil {
						fmt.Println("Write error : ", err)
					}
				}(recepient)
			} else {
				//store it to the database
				s.handlePostNotifcation(notification)
			}
		}
		if broad.Type == "request" {
			for k, v := range s.active_conns {
				if v == ws {
					fmt.Println("Gotcha", k)
					message := &Message{
						ID:         uuid.New(),
						Content:    broad.Content, //holds the uuid of the book
						SenderID:   uuid.MustParse(k),
						ReceiverID: uuid.MustParse(broad.ID),
						SentAt:     time.Now(),
						Seen:       false,
						Request:    true,
					}
					s.broadcast_message(message, ws)
				}
			}

		}
		defer s.remove(ws, broad.ID)
	}
}

func (s *ApiHandler) remove(ws *websocket.Conn, id string) {
	delete(s.active_conns, id)
	s.broadcast_to_all(ws)
}

func (s *ApiHandler) broadcast_to_all(ws *websocket.Conn) {
	data := []*broad_id{}
	for id, _ := range s.active_conns {
		broad := broad_id{
			ID: id,
		}
		data = append(data, &broad)
	}
	message := return_message{
		Type:  "active",
		Conns: data,
	}
	marshal, err := json.Marshal(message)
	if err != nil {
		fmt.Println("There is an error in converting to json", err)
	}
	for _, aws := range s.active_conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(marshal); err != nil {
				fmt.Println("Write error : ", err)
			}
		}(aws)
	}
}

// Search endpoint
func (s *ApiHandler) SearchWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		id := uuid.MustParse(r.URL.Query().Get("id"))
		fmt.Println("New search incoming connection from client : ", ws.RemoteAddr())
		s.search(ws, context.Background(), id.String())
	})
	wsHandler.ServeHTTP(w, r)
}

func (s *ApiHandler) search(ws *websocket.Conn, ctx context.Context, id string) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Read error is enconuntered :", err)
			continue
		}
		msg := buf[:n]
		query := string(msg)
		//Fetch from the db alike to query
		fmt.Println(query)
		accounts, err := s.SearchAccount(ctx, s.db, query, id)
		if err != nil {
			fmt.Println("Error in the database pull function", err)
		}
		//converting the message to json
		marshal_message, erro := json.Marshal(accounts)
		if erro != nil {
			fmt.Println("Error while marshalling the returning account object ", erro)
		}
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(marshal_message); err != nil {
				fmt.Println("Write error : ", err)
			}
		}(ws)
	}
}
