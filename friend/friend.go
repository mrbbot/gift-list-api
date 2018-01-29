package friend

import (
	"net/http"
	"database/sql"
	"../util"
	"encoding/json"
	"github.com/gorilla/mux"
)

type Friend struct {
	ID 		int64 	`json:"id"`
	Owner 	string 	`json:"owner"`
	Friend 	string 	`json:"friend"`
	State 	bool 	`json:"state"`
}

type emailContainer struct {
	Email	string 	`json:"email"`
}

func GetFriends(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	friends := []Friend{}

	//TODO: Add authorisation
	owner := "Tester"

	rows, err := db.Query("SELECT * FROM friends WHERE owner = ?", owner)
	if err != nil {
		util.EncodeError(w, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var friend Friend
		err := rows.Scan(&friend.ID, &friend.Owner, &friend.Friend, &friend.State)
		if err != nil {
			util.EncodeError(w, err)
			return
		}

		friends = append(friends, friend)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(friends)
}

func AddFriend(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var email emailContainer
	json.NewDecoder(r.Body).Decode(&email)

	//TODO: Add authorisation
	owner := "Tester"

	friend := Friend{Owner: owner, Friend: email.Email, State: false}

	res, err := db.Exec("INSERT INTO friends (owner, friend, state) VALUES (?, ?, ?)", friend.Owner, friend.Friend, friend.State)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	friend.ID, err = res.LastInsertId()
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(friend)
}

func AcceptFriend(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	params := mux.Vars(r)
	friendId := params["friendId"]

	var currentFriend Friend
	err := db.QueryRow("SELECT * FROM friends WHERE id = ?", friendId).Scan(
		&currentFriend.ID, &currentFriend.Owner, &currentFriend.Friend, &currentFriend.State)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	//TODO: Add authorisation
	friend := "me@mrbbot.co.uk"

	if (currentFriend.Friend != friend) || (currentFriend.State) {
		util.EncodeUnauthorised(w)
		return
	}

	currentFriend.State = true
	_, err = db.Exec("UPDATE friends SET state = ? WHERE id = ?", currentFriend.State, friendId)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	_, err = db.Exec("INSERT INTO friends (owner, friend, state) VALUES (?, ?, ?)", currentFriend.Friend, currentFriend.Owner, 1)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentFriend)
}

func RemoveFriend(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	params := mux.Vars(r)
	id := params["friendId"]

	res, err := db.Exec("DELETE FROM friends WHERE id = ?", id)
	if err != nil {
		util.EncodeError(w, err)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if rowsAffected > 0 {
		json.NewEncoder(w).Encode(util.Response{Success: true})
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(util.Response{Success: false, Message: "friend not found"})
	}
}