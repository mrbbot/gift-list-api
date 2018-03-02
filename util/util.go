package util

import (
	"net/http"
	"fmt"
	"encoding/json"
	"log"
)

type Response struct {
	Success		bool		`json:"success"`
	Message 	string		`json:"message,omitempty"`
}

func EncodeError(w http.ResponseWriter, err error) {
	log.Printf("\tERROR: %v", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(Response{Success: false, Message: fmt.Sprintf("%v", err)})
}

func EncodeUnauthorised(w http.ResponseWriter) {
	log.Printf("\tUNAUTHORISED")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(Response{Success: false, Message: "unauthorised"})
}

func EncodeNotFound(w http.ResponseWriter) {
	log.Printf("\tNOT FOUND")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(Response{Success: false, Message: "not found"})
}