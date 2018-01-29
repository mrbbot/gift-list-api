package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"./list"
	"./gift"
	"./friend"
)

//TODO: Consider optimising with prepared statements
//https://newfivefour.com/golang-closures-anonymous-functions.html
//https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702
func main() {
	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/giftlist")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	injectDb := func(f func(http.ResponseWriter, *http.Request, *sql.DB)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			f(w, r, db)
		}
	}

	router := mux.NewRouter()

	router.HandleFunc("/lists/{userId}", injectDb(list.GetLists)).Methods("GET")
	router.HandleFunc("/list", injectDb(list.CreateList)).Methods("POST")
	router.HandleFunc("/list/{listId}", injectDb(list.EditList)).Methods("POST")
	router.HandleFunc("/list/{listId}", injectDb(list.RemoveList)).Methods("DELETE")

	router.HandleFunc("/list/{listId}/gift", injectDb(gift.CreateGift)).Methods("POST")
	router.HandleFunc("/list/{listId}/gift/{giftId}", injectDb(gift.EditGift)).Methods("POST")
	router.HandleFunc("/list/{listId}/gift/{giftId}", injectDb(gift.RemoveGift)).Methods("DELETE")
	router.HandleFunc("/list/{listId}/gift/{giftId}/claim", injectDb(gift.ClaimGift)).Methods("POST")

	router.HandleFunc("/friends", injectDb(friend.GetFriends)).Methods("GET")
	router.HandleFunc("/friend", injectDb(friend.AddFriend)).Methods("POST")
	router.HandleFunc("/friend/accept/{friendId}", injectDb(friend.AcceptFriend)).Methods("POST")
	router.HandleFunc("/friend/{friendId}", injectDb(friend.RemoveFriend)).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}