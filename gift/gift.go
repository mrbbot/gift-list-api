package gift

import (
	"net/http"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"../util"
)

type Gift struct {
	ID			int64	`json:"id"`
	Name 		string	`json:"name"`
	Description	string	`json:"description"`
	Url 		string	`json:"url"`
	ImageUrl 	string	`json:"imageUrl"`
	Claim		*Claim 	`json:"claim"`
}

type Claim struct {
	State		int 	`json:"state"`
	User		string 	`json:"user"`
}

func CreateGift(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	params := mux.Vars(r)

	var gift Gift
	json.NewDecoder(r.Body).Decode(&gift)
	gift.Claim = &Claim{State: 0, User: ""}

	res, err := db.Exec("INSERT INTO gifts (name, description, url, image_url, list_id, claim_status, claimed_by) VALUES (?, ?, ?, ?, ?, ?, ?)",
		gift.Name, gift.Description, gift.Url, gift.ImageUrl, params["listId"], gift.Claim.State, gift.Claim.User)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	gift.ID, err = res.LastInsertId()
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&gift)

}

func EditGift(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	params := mux.Vars(r)
	giftId := params["giftId"]

	var currentGift Gift
	currentGift.Claim = &Claim{}
	err := db.QueryRow("SELECT id, name, description, url, image_url, claim_status, claimed_by FROM gifts WHERE id = ?", giftId).Scan(
		&currentGift.ID, &currentGift.Name, &currentGift.Description, &currentGift.Url, &currentGift.ImageUrl, &currentGift.Claim.State, &currentGift.Claim.User)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	var newGift Gift
	json.NewDecoder(r.Body).Decode(&newGift)
	if len(newGift.Name) > 0 {
		currentGift.Name = newGift.Name
	}
	if len(newGift.Description) > 0 {
		currentGift.Description = newGift.Description
	}
	if len(newGift.Url) > 0 {
		currentGift.Url = newGift.Url
	}
	if len(newGift.ImageUrl) > 0 {
		currentGift.ImageUrl = newGift.ImageUrl
	}

	_, err = db.Exec("UPDATE gifts SET name = ?, description = ?, url = ?, image_url = ? WHERE id = ?", currentGift.Name, currentGift.Description, currentGift.Url, currentGift.ImageUrl, giftId)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentGift)
}

func RemoveGift(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	params := mux.Vars(r)
	giftId := params["giftId"]

	res, err := db.Exec("DELETE FROM gifts WHERE id = ?", giftId)
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
		json.NewEncoder(w).Encode(util.Response{Success: false, Message: "gift not found"})
	}
}

func ClaimGift(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	params := mux.Vars(r)
	giftId := params["giftId"]

	var claim Claim
	err := db.QueryRow("SELECT claim_status, claimed_by FROM gifts WHERE id = ?", giftId).Scan(
		&claim.State, &claim.User)
	if err != nil {
		util.EncodeError(w, err)
		return
	}

	// TODO: Get from authorisation
	claimee := "Claimer"

	if (len(claim.User) == 0) || (claim.User == claimee) {
		var newClaim Claim
		json.NewDecoder(r.Body).Decode(&newClaim)

		claim.User = claimee
		claim.State = newClaim.State

		if claim.State == 0 {
			claim.User = ""
		}

		_, err := db.Exec("UPDATE gifts SET claim_status = ?, claimed_by = ? WHERE id = ?", claim.State, claim.User, giftId)
		if err != nil {
			util.EncodeError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claim)
	} else {
		util.EncodeUnauthorised(w)
	}
}