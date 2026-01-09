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

type Account struct {
	ID             int     `json:"id"`
	UserID         int     `json:"user_id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Currency       string  `json:"currency"`
	InitialBalance float64 `json:"initial_balance"`
	CurrentBalance float64 `json:"current_balance"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
}

type Transaction struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	AccountID   int     `json:"account_id"`
	CategoryID  int     `json:"category_id"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Notes       string  `json:"notes"`
	CreatedAt   string  `json:"created_at"`
}

type Category struct {
	ID     int    `json:"id"`
	UserID *int   `json:"user_id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
}

type Budget struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	CategoryID  int     `json:"category_id"`
	LimitAmount float64 `json:"limit_amount"`
	Period      string  `json:"period"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	CreatedAt   string  `json:"created_at"`
}

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
	);

	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		currency TEXT DEFAULT 'USD',
		initial_balance REAL DEFAULT 0,
		current_balance REAL DEFAULT 0,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		icon TEXT,
		color TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		account_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		amount REAL NOT NULL,
		type TEXT NOT NULL,
		date DATETIME NOT NULL,
		description TEXT,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (account_id) REFERENCES accounts(id),
		FOREIGN KEY (category_id) REFERENCES categories(id)
	);

	CREATE TABLE IF NOT EXISTS budgets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		limit_amount REAL NOT NULL,
		period TEXT NOT NULL,
		start_date DATETIME NOT NULL,
		end_date DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (category_id) REFERENCES categories(id)
	)`

	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	// Insert default categories
	defaultCategories := `INSERT OR IGNORE INTO categories (user_id, name, type, icon, color) VALUES
		(NULL, 'Salary', 'income', '💼', '#10b981'),
		(NULL, 'Freelance', 'income', '💻', '#3b82f6'),
		(NULL, 'Investment', 'income', '📈', '#8b5cf6'),
		(NULL, 'Food & Dining', 'expense', '🍔', '#ef4444'),
		(NULL, 'Transportation', 'expense', '🚗', '#f59e0b'),
		(NULL, 'Shopping', 'expense', '🛍️', '#ec4899'),
		(NULL, 'Entertainment', 'expense', '🎬', '#a855f7'),
		(NULL, 'Bills & Utilities', 'expense', '📄', '#14b8a6'),
		(NULL, 'Healthcare', 'expense', '🏥', '#06b6d4'),
		(NULL, 'Education', 'expense', '📚', '#3b82f6')`

	_, err = db.Exec(defaultCategories)
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

// Account handlers
func getAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT id, user_id, name, type, currency, initial_balance, current_balance, is_active, created_at FROM accounts WHERE user_id = ? AND is_active = 1", userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		rows.Scan(&acc.ID, &acc.UserID, &acc.Name, &acc.Type, &acc.Currency, &acc.InitialBalance, &acc.CurrentBalance, &acc.IsActive, &acc.CreatedAt)
		accounts = append(accounts, acc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	var acc Account
	if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if acc.Currency == "" {
		acc.Currency = "USD"
	}
	acc.CurrentBalance = acc.InitialBalance

	result, err := db.Exec("INSERT INTO accounts (user_id, name, type, currency, initial_balance, current_balance) VALUES (?, ?, ?, ?, ?, ?)",
		acc.UserID, acc.Name, acc.Type, acc.Currency, acc.InitialBalance, acc.CurrentBalance)
	if err != nil {
		http.Error(w, "Error creating account", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	acc.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(acc)
}

// Transaction handlers
func getTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	accountID := r.URL.Query().Get("account_id")

	query := "SELECT id, user_id, account_id, category_id, amount, type, date, description, notes, created_at FROM transactions WHERE user_id = ?"
	args := []interface{}{userID}

	if accountID != "" {
		query += " AND account_id = ?"
		args = append(args, accountID)
	}
	query += " ORDER BY date DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var txn Transaction
		rows.Scan(&txn.ID, &txn.UserID, &txn.AccountID, &txn.CategoryID, &txn.Amount, &txn.Type, &txn.Date, &txn.Description, &txn.Notes, &txn.CreatedAt)
		transactions = append(transactions, txn)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	var txn Transaction
	if err := json.NewDecoder(r.Body).Decode(&txn); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert transaction
	result, err := tx.Exec("INSERT INTO transactions (user_id, account_id, category_id, amount, type, date, description, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		txn.UserID, txn.AccountID, txn.CategoryID, txn.Amount, txn.Type, txn.Date, txn.Description, txn.Notes)
	if err != nil {
		http.Error(w, "Error creating transaction", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	txn.ID = int(id)

	// Update account balance
	balanceChange := txn.Amount
	if txn.Type == "debit" {
		balanceChange = -balanceChange
	}

	_, err = tx.Exec("UPDATE accounts SET current_balance = current_balance + ? WHERE id = ?", balanceChange, txn.AccountID)
	if err != nil {
		http.Error(w, "Error updating balance", http.StatusInternalServerError)
		return
	}

	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(txn)
}

// Category handlers
func getCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, user_id, name, type, icon, color FROM categories")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.Icon, &cat.Color)
		categories = append(categories, cat)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// Budget handlers
func getBudgets(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	rows, err := db.Query("SELECT id, user_id, category_id, limit_amount, period, start_date, end_date, created_at FROM budgets WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var budget Budget
		rows.Scan(&budget.ID, &budget.UserID, &budget.CategoryID, &budget.LimitAmount, &budget.Period, &budget.StartDate, &budget.EndDate, &budget.CreatedAt)
		budgets = append(budgets, budget)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

func createBudget(w http.ResponseWriter, r *http.Request) {
	var budget Budget
	if err := json.NewDecoder(r.Body).Decode(&budget); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO budgets (user_id, category_id, limit_amount, period, start_date, end_date) VALUES (?, ?, ?, ?, ?, ?)",
		budget.UserID, budget.CategoryID, budget.LimitAmount, budget.Period, budget.StartDate, budget.EndDate)
	if err != nil {
		http.Error(w, "Error creating budget", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	budget.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(budget)
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Database error:", err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.Use(cors)

	// Auth routes
	router.HandleFunc("/api/auth/register", register).Methods("POST")
	router.HandleFunc("/api/auth/login", login).Methods("POST")

	// Account routes
	router.HandleFunc("/api/accounts", getAccounts).Methods("GET")
	router.HandleFunc("/api/accounts", createAccount).Methods("POST")

	// Transaction routes
	router.HandleFunc("/api/transactions", getTransactions).Methods("GET")
	router.HandleFunc("/api/transactions", createTransaction).Methods("POST")

	// Category routes
	router.HandleFunc("/api/categories", getCategories).Methods("GET")

	// Budget routes
	router.HandleFunc("/api/budgets", getBudgets).Methods("GET")
	router.HandleFunc("/api/budgets", createBudget).Methods("POST")

	log.Println("Finance Manager API running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
