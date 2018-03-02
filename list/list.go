package list

import (
	"net/http"
	"../gift"
	"../util"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"firebase.google.com/go/auth"
)

type List struct {
	ID			int64			`json:"id"`
	Name		string 			`json:"name"`
	Owner		string 			`json:"owner"`
	Description string 			`json:"description"`
	Gifts		[]*gift.Gift	`json:"gifts"`
}

func getListGifts(db *sql.DB, listId int64) ([]*gift.Gift, error) {
	gifts := []*gift.Gift{}

	rows, err := db.Query("SELECT gifts.id, gifts.name, gifts.description, gifts.url, gifts.image_url, gifts.claim_status, gifts.claimed_by FROM lists, gifts WHERE lists.id = gifts.list_id AND lists.id = ?", listId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g gift.Gift
		g.Claim = &gift.Claim{}
		err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Url, &g.ImageUrl, &g.Claim.State, &g.Claim.User)
		if err != nil {
			return nil, err
		}

		gifts = append(gifts, &g)
	}

	return gifts, nil
}

func GetLists(w http.ResponseWriter, r *http.Request, db *sql.DB, user *auth.Token) {
	params := mux.Vars(r)

	lists := []List{}

	rows, err := db.Query("SELECT * FROM lists WHERE owner = ?", params["userId"])
	if err != nil {
		util.EncodeError(w, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var list List
		err := rows.Scan(&list.ID, &list.Name, &list.Owner, &list.Description)
		if err != nil {
			util.EncodeError(w, err)
			return
		}

		list.Gifts, err = getListGifts(db, list.ID)
		if err != nil {
			util.EncodeError(w, err)
			return
		}

		lists = append(lists, list)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

func CreateList(w http.ResponseWriter, r *http.Request, db *sql.DB, user *auth.Token) {
	var list List
	json.NewDecoder(r.Body).Decode(&list)
	list.Owner = "Tester" //TODO: Add Authentication

	res, err := db.Exec("INSERT INTO lists (name, owner, description) VALUES (?, ?, ?)", list.Name, list.Owner, list.Description)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	list.ID, err = res.LastInsertId()
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	list.Gifts = []*gift.Gift{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func EditList(w http.ResponseWriter, r *http.Request, db *sql.DB, user *auth.Token) {
	params := mux.Vars(r)
	id := params["listId"]

	var currentList List
	err := db.QueryRow("SELECT * FROM lists WHERE id = ?", id).Scan(&currentList.ID, &currentList.Name, &currentList.Owner, &currentList.Description)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	var newList List
	json.NewDecoder(r.Body).Decode(&newList)
	if len(newList.Name) > 0 {
		currentList.Name = newList.Name
	}
	if len(newList.Description) > 0 {
		currentList.Description = newList.Description
	}

	_, err = db.Exec("UPDATE lists SET name = ?, description = ? WHERE id = ?", currentList.Name, currentList.Description, id)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentList)
}

func RemoveList(w http.ResponseWriter, r *http.Request, db *sql.DB, user *auth.Token) {
	params := mux.Vars(r)
	id := params["listId"]

	res, err := db.Exec("DELETE FROM lists WHERE id = ?", id)
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
		json.NewEncoder(w).Encode(util.Response{Success: false, Message: "list not found"})
	}
}