package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./finance.db")
	if err != nil {
		return err
	}

	schema := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.Exec(schema)
	return err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func register(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	json.NewDecoder(r.Body).Decode(&req)

	hash, _ := hashPassword(req["password"])
	result, err := db.Exec("INSERT INTO users (email, password_hash, first_name, last_name) VALUES (?, ?, ?, ?)",
		req["email"], hash, req["first_name"], req["last_name"])

	if err != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	id, _ := result.LastInsertId()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registration successful",
		"user_id": id,
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	json.NewDecoder(r.Body).Decode(&req)

	var id int
	var passwordHash, firstName, lastName string
	err := db.QueryRow("SELECT id, password_hash, first_name, last_name FROM users WHERE email = ?", req["email"]).
		Scan(&id, &passwordHash, &firstName, &lastName)

	if err != nil || !checkPasswordHash(req["password"], passwordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Login successful",
		"user_id":    id,
		"first_name": firstName,
		"last_name":  lastName,
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Database error:", err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.Use(cors)

	router.HandleFunc("/api/auth/register", register).Methods("POST")
	router.HandleFunc("/api/auth/login", login).Methods("POST")

	log.Println("Finance Manager API running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
