package friend

import (
	"net/http"
	"database/sql"
	"../util"
	"encoding/json"
	authHelper "../auth"
	"github.com/gorilla/mux"
	"firebase.google.com/go/auth"
	"strconv"
)

type Friend struct {
	ID 		int64 	`json:"id"`
	Owner 	string 	`json:"owner,omitempty"`
	Friend 	string 	`json:"friend,omitempty"`
	Name	string 	`json:"name,omitempty"`
	State 	bool 	`json:"state"`
}

type emailContainer struct {
	Email	string 	`json:"email"`
}

type friendContainer struct {
	Current []Friend `json:"current"`
	Requests []Friend `json:"requests"`
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

func GetFriends(w http.ResponseWriter, r *http.Request, db *sql.DB, user *auth.Token) {
	var container friendContainer

	container.Current = []Friend{}
	currentRows, err := db.Query("SELECT id, friend, state FROM friends WHERE owner = ?", user.UID)
	if err != nil {
		util.EncodeError(w, err)
		return
	}
	defer currentRows.Close()
	for currentRows.Next() {
		var friend Friend
		err := currentRows.Scan(&friend.ID, &friend.Friend, &friend.State)
		if err != nil {
			util.EncodeError(w, err)
			return
		}

		user, err := authHelper.UserFromUID(friend.Friend)
		if err != nil {
			util.EncodeError(w, err)
			return
		}
		friend.Friend = user.Email
		friend.Name = user.DisplayName

		container.Current = append(container.Current, friend)
	}

	container.Requests = []Friend{}
	requestRows, err := db.Query("SELECT id, owner, state FROM friends WHERE friend = ? AND state = 0", user.UID)
	if err != nil {
		util.EncodeError(w, err)
		return
	}
	defer requestRows.Close()
	for requestRows.Next() {
		var friend Friend
		err := requestRows.Scan(&friend.ID, &friend.Owner, &friend.State)
		if err != nil {
			util.EncodeError(w, err)
			return
		}

		user, err := authHelper.UserFromUID(friend.Owner)
		if err != nil {
			util.EncodeError(w, err)
			return
		}
		friend.Owner = user.Email
		friend.Name = user.DisplayName

		container.Requests = append(container.Requests, friend)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(container)
}

func AddFriend(w http.ResponseWriter, r *http.Request, db *sql.DB, user *auth.Token) {
	var email emailContainer
	json.NewDecoder(r.Body).Decode(&email)

	friendUser, err := authHelper.UserFromEmail(email.Email)
	if err != nil {
		util.EncodeNotFound(w)
		return
	}

	friend := Friend{Owner: user.UID, Friend: friendUser.UID, State: false}
	if user.UID == friendUser.UID {
		util.EncodeUnauthorised(w)
		return
	}

	// Check if the users are already friends
	existingFriend, err := db.Query("SELECT id FROM friends WHERE owner = ? AND friend = ?", user.UID, friendUser.UID)
	if err != nil {
		util.EncodeError(w, err)
		return
	}
	defer existingFriend.Close()
	if existingFriend.Next() {
		util.EncodeUnauthorised(w)
		return
	}

	// Check if there is a pending friend request the other way
	existingFriendRequest, err := db.Query("SELECT id, owner, friend FROM friends WHERE owner = ? AND friend = ? AND state = 0", friendUser.UID, user.UID)
	if err != nil {
		util.EncodeError(w, err)
		return
	}
	defer existingFriendRequest.Close()
	if existingFriendRequest.Next() {
		var (
			friendId int64
			ownerUid string
			friendUid string
		)
		existingFriendRequest.Scan(&friendId, &ownerUid, &friendUid)
		doAcceptFriend(db, strconv.FormatInt(friendId, 10), ownerUid, friendUid)

		w.Header().Set("Content-Type", "application/json")
		friend.ID = friendId
		friend.State = true
		json.NewEncoder(w).Encode(friend)

		return
	}

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

func doAcceptFriend(db *sql.DB, friendId string, owner string, friend string) error {
	_, err := db.Exec("UPDATE friends SET state = ? WHERE id = ?", true, friendId)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO friends (owner, friend, state) VALUES (?, ?, ?)", friend, owner, true)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Consider removing this method and just use the add friend route instead
func AcceptFriend(w http.ResponseWriter, r *http.Request, db *sql.DB, user *auth.Token) {
	params := mux.Vars(r)
	friendId := params["friendId"]

	var currentFriend Friend
	err := db.QueryRow("SELECT * FROM friends WHERE id = ?", friendId).Scan(
		&currentFriend.ID, &currentFriend.Owner, &currentFriend.Friend, &currentFriend.State)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	if (currentFriend.Friend != user.UID) || (currentFriend.State) {
		util.EncodeUnauthorised(w)
		return
	}

	currentFriend.State = true
	err = doAcceptFriend(db, friendId, currentFriend.Owner, currentFriend.Friend)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentFriend)
}

func RemoveFriend(w http.ResponseWriter, r *http.Request, db *sql.DB, user *auth.Token) {
	params := mux.Vars(r)
	id := params["friendId"]

	var currentFriend Friend
	err := db.QueryRow("SELECT * FROM friends WHERE id = ?", id).Scan(
		&currentFriend.ID, &currentFriend.Owner, &currentFriend.Friend, &currentFriend.State)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	if (currentFriend.Owner != user.UID) && (currentFriend.Friend != user.UID) {
		util.EncodeUnauthorised(w)
		return
	}

	res, err := db.Exec("DELETE FROM friends WHERE (owner = ? AND friend = ?) OR (owner = ? AND friend = ?)", currentFriend.Owner, currentFriend.Friend, currentFriend.Friend, currentFriend.Owner)
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