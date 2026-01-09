package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	db        *sql.DB
	jwtSecret = []byte("your-secret-key-change-in-production")
)

type User struct {
	ID                 int       `+"`json:\"id\"`"+`
	Email              string    `+"`json:\"email\"`"+`
	PasswordHash       string    `+"`json:\"-\"`"+`
	FirstName          string    `+"`json:\"first_name\"`"+`
	LastName           string    `+"`json:\"last_name\"`"+`
	CurrencyPreference string    `+"`json:\"currency_preference\"`"+`
	Role               string    `+"`json:\"role\"`"+`
	CreatedAt          time.Time `+"`json:\"created_at\"`"+`
}

type Account struct {
	ID             int       `+"`json:\"id\"`"+`
	UserID         int       `+"`json:\"user_id\"`"+`
	Name           string    `+"`json:\"name\"`"+`
	Type           string    `+"`json:\"type\"`"+`
	Currency       string    `+"`json:\"currency\"`"+`
	InitialBalance float64   `+"`json:\"initial_balance\"`"+`
	CurrentBalance float64   `+"`json:\"current_balance\"`"+`
	IsActive       bool      `+"`json:\"is_active\"`"+`
	CreatedAt      time.Time `+"`json:\"created_at\"`"+`
}

type Claims struct {
	UserID int    `+"`json:\"user_id\"`"+`
	Email  string `+"`json:\"email\"`"+`
	Role   string `+"`json:\"role\"`"+`
	jwt.RegisteredClaims
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./finance.db")
	if err != nil {
		return err
	}

	schema := `+"`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, first_name TEXT NOT NULL, last_name TEXT NOT NULL, currency_preference TEXT DEFAULT 'USD', role TEXT DEFAULT 'user', created_at DATETIME DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS accounts (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, name TEXT NOT NULL, type TEXT NOT NULL, currency TEXT DEFAULT 'USD', initial_balance REAL DEFAULT 0, current_balance REAL DEFAULT 0, is_active BOOLEAN DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (user_id) REFERENCES users(id)); CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, name TEXT NOT NULL, type TEXT NOT NULL, icon TEXT, color TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, account_id INTEGER NOT NULL, category_id INTEGER NOT NULL, amount REAL NOT NULL, type TEXT NOT NULL, date DATETIME NOT NULL, description TEXT, notes TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (user_id) REFERENCES users(id), FOREIGN KEY (account_id) REFERENCES accounts(id), FOREIGN KEY (category_id) REFERENCES categories(id));`+"`"+`

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

func generateJWT(userID int, email, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `+"`json:\"email\"`"+`
		Password  string `+"`json:\"password\"`"+`
		FirstName string `+"`json:\"first_name\"`"+`
		LastName  string `+"`json:\"last_name\"`"+`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hash, _ := hashPassword(req.Password)
	result, err := db.Exec("INSERT INTO users (email, password_hash, first_name, last_name) VALUES (?, ?, ?, ?)", req.Email, hash, req.FirstName, req.LastName)
	if err != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	id, _ := result.LastInsertId()
	token, _ := generateJWT(int(id), req.Email, "user")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"token": token, "user": map[string]interface{}{"id": id, "email": req.Email, "first_name": req.FirstName, "last_name": req.LastName}})
}

func login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `+"`json:\"email\"`"+`
		Password string `+"`json:\"password\"`"+`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow("SELECT id, email, password_hash, first_name, last_name, role FROM users WHERE email = ?", req.Email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role)

	if err != nil || !checkPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := generateJWT(user.ID, user.Email, user.Role)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"token": token, "user": user})
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
		log.Fatal("Database initialization failed:", err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.Use(cors)

	router.HandleFunc("/api/auth/register", register).Methods("POST")
	router.HandleFunc("/api/auth/login", login).Methods("POST")

	log.Println("🚀 Finance Manager API running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
