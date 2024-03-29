package api

//Handler for the /community endpoint. 

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type CreateCommunity struct {
	CommunityName string `json:"community_name"`
}

//Handling creating a community. 
func (s *ApiHandler) handleCreateCommunity(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	createCommunity := CreateCommunity{}
	err := json.NewDecoder(r.Body).Decode(&createCommunity)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(createCommunity.CommunityName)
  //Perform a check to see whetehr the community name already exists in the database. 
	checkCommunity := Community{}
	s.db.First(&checkCommunity, "community_name = ?", createCommunity.CommunityName)
	if checkCommunity.ID.String() == null_uuid {
    //The community name does not exist in the database.
		fmt.Println("Community name is unique ")
		community := s.NewCommunity(createCommunity.CommunityName)
		s.db.Create(community)
		return s.WriteJson(w, http.StatusOK, community)
	} else {
    //The community name already exists in the database.
		return s.WriteJson(w, http.StatusBadRequest, "Found another community with the same name")
	}
}

func (s *ApiHandler) handleGetAllComms(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	response, err := s.GetAllComms(ctx, s.db)
	if err != nil {
		fmt.Println(err.Error())
	}
	return s.WriteJson(w, http.StatusOK, response)
}

func (s *ApiHandler) handleGetCommunity(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	comm := Community{}
	s.db.First(&comm, "id = ?", id)
	return s.WriteJson(w, http.StatusOK, comm)
}
