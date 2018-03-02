package util

import (
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"database/sql"
)

type Response struct {
	Success		bool		`json:"success"`
	Message 	string		`json:"message,omitempty"`
}

func EncodeError(w http.ResponseWriter, err error) {
	log.Printf("\t<- ERROR: %v", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(Response{Success: false, Message: fmt.Sprintf("%v", err)})
}

func EncodeUnauthorised(w http.ResponseWriter) {
	log.Printf("\t<- UNAUTHORISED")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(Response{Success: false, Message: "unauthorised"})
}

func EncodeNotFound(w http.ResponseWriter) {
	log.Printf("\t<- NOT FOUND")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(Response{Success: false, Message: "not found"})
}

func GetListOwner(db *sql.DB, listId string) (string, error) {
	var currentOwner string
	err := db.QueryRow("SELECT owner FROM lists WHERE id = ?", listId).Scan(&currentOwner)
	if err != nil {
		return "", err
	}
	return currentOwner, nil
}

func AreFriends(db *sql.DB, uidOne string, uidTwo string) (bool, error) {
	if uidOne == uidTwo {
		return true, nil
	}
	rows, err := db.Query("SELECT id FROM friends WHERE owner = ? AND friend = ? AND state = 1", uidOne, uidTwo)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}
