package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

type Message_Posted struct {
	Content     string `json: "content"`
	Sender_ID   string `json: "sender_id"`
	Receiver_ID string `json: "receiver_id"`
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
		ws.Write([]byte("thank you for the msg!!! "))
		message := &Message{
			ID:         uuid.New(),
			Content:    post_message.Content,
			SenderID:   uuid.MustParse(post_message.Sender_ID),
			ReceiverID: uuid.MustParse(post_message.Receiver_ID),
			SentAt:     time.Now(),
		}
		s.broadcast_message(message)
	}
}

func (s *ApiHandler) broadcast_message(message *Message) {
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
	//Need to write the message to the database

}