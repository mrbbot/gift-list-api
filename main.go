package main

import (
	authHelper "github.com/mrbbot/gift-list-api/auth"
	"github.com/mrbbot/gift-list-api/friend"
	"github.com/mrbbot/gift-list-api/gift"
	"github.com/mrbbot/gift-list-api/list"
	"github.com/mrbbot/gift-list-api/util"
	"database/sql"
	"firebase.google.com/go/auth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"strconv"
)

//TODO: Consider optimising with prepared statements
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v\n", err)
	}

	db, err := sql.Open("mysql", os.Getenv("DB"))
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
	router.HandleFunc("/friend/reject/{friendId}", inject(friend.RejectFriend)).Methods("POST")
	router.HandleFunc("/friend/{friendId}", inject(friend.RemoveFriend)).Methods("DELETE")

	handler := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		AllowedMethods: []string{"GET", "POST", "DELETE"},
	}).Handler(router)

	address := os.Getenv("ADDRESS")
	useSSL, err := strconv.ParseBool(os.Getenv("SSL"))
	if useSSL {
		certPath := os.Getenv("CERT")
		keyPath := os.Getenv("KEY")
		log.Fatal(http.ListenAndServeTLS(address, certPath, keyPath, handler))
	} else {
		log.Fatal(http.ListenAndServe(address, handler))
	}
}
