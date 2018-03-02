package main

import (
	"./friend"
	"./gift"
	"./list"
	"./util"
	authHelper "./auth"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"firebase.google.com/go/auth"
)

const address = ":8081"

//TODO: Consider optimising with prepared statements
func main() {
	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/giftlist")
	if err != nil {
		log.Fatalf("error initializing database: %v\n", err)
	}
	defer db.Close()

	authHelper.Init()

	inject := func(f func(http.ResponseWriter, *http.Request, *sql.DB, *auth.Token)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			token, err := authHelper.Verify(r.Header.Get("Authorization"))

			if err != nil {
				log.Printf("<UNAUTHORISED> (%v) -> [%s] %v\n", err, r.Method, r.RequestURI)
				util.EncodeUnauthorised(w)
				return
			}

			log.Printf("%s -> [%s] %v\n", token.UID, r.Method, r.RequestURI)
			f(w, r, db, token)
		}
	}

	router := mux.NewRouter()

	router.HandleFunc("/lists/{userId}", inject(list.GetLists)).Methods("GET")
	router.HandleFunc("/list", inject(list.CreateList)).Methods("POST")
	router.HandleFunc("/list/{listId}", inject(list.EditList)).Methods("POST")
	router.HandleFunc("/list/{listId}", inject(list.RemoveList)).Methods("DELETE")

	router.HandleFunc("/list/{listId}/gift", inject(gift.CreateGift)).Methods("POST")
	router.HandleFunc("/list/{listId}/gift/{giftId}", inject(gift.EditGift)).Methods("POST")
	router.HandleFunc("/list/{listId}/gift/{giftId}", inject(gift.RemoveGift)).Methods("DELETE")
	router.HandleFunc("/list/{listId}/gift/{giftId}/claim", inject(gift.ClaimGift)).Methods("POST")

	router.HandleFunc("/friends", inject(friend.GetFriends)).Methods("GET")
	router.HandleFunc("/friend", inject(friend.AddFriend)).Methods("POST")
	router.HandleFunc("/friend/accept/{friendId}", inject(friend.AcceptFriend)).Methods("POST")
	router.HandleFunc("/friend/{friendId}", inject(friend.RemoveFriend)).Methods("DELETE")

	log.Fatal(http.ListenAndServe(address, router))
}
