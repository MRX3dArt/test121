package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

type Client struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "postgresql://user:password@localhost/bank?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println(db)

	router := mux.NewRouter()
	router.HandleFunc("/clients/{id}", getClient).Methods("GET")
	router.HandleFunc("/clients/transfer", transfer).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getClient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	client := Client{}
	err := db.QueryRow("SELECT id, name, balance FROM clients WHERE id = $1", id).Scan(&client.ID, &client.Name, &client.Balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(client)
}

func transfer(w http.ResponseWriter, r *http.Request) {
	var transfer struct {
		SenderID   int     `json:"sender_id"`
		ReceiverID int     `json:"receiver_id"`
		Amount     float64 `json:"amount"`
	}
	err := json.NewDecoder(r.Body).Decode(&transfer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var senderBalance float64
	err = tx.QueryRow("SELECT balance FROM clients WHERE id = $1", transfer.SenderID).Scan(&senderBalance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if senderBalance < transfer.Amount {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec("UPDATE clients SET balance = balance - $1 WHERE id = $2", transfer.Amount, transfer.SenderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = tx.Exec("UPDATE clients SET balance = balance + $1 WHERE id = $2", transfer.Amount, transfer.ReceiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Transfer successful")
}
