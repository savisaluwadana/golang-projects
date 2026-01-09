# Finance Manager Backend API

A robust financial management backend built with Go, featuring user authentication, account management, transaction tracking, budgeting, and financial reporting.

## 🚀 Features Implemented

### ✅ Core Features
- **User Authentication**: Secure registration and login with bcrypt password hashing
- **Account Management**: Create and manage multiple accounts (checking, savings, credit cards, cash, investments)
- **Transaction Tracking**: Record income and expenses with automatic balance updates
- **Categories**: Pre-loaded income/expense categories with custom support
- **Budgeting System**: Set spending limits per category with period tracking
- **Database**: SQLite database with proper schema and foreign key constraints
- **CORS Support**: Cross-origin requests enabled for frontend integration
- **RESTful API**: Clean API endpoints following REST principles
- **Atomic Transactions**: Balance updates are transactional to prevent data inconsistency

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Database**: SQLite3
- **Router**: Gorilla Mux
- **Password Hashing**: bcrypt
- **Dependencies**:
  - `github.com/gorilla/mux` - HTTP router
  - `github.com/mattn/go-sqlite3` - SQLite driver
  - `golang.org/x/crypto/bcrypt` - Password hashing

## 📂 Project Structure

```
Finance manager application backend/
├── server.go           # Main application server
├── go.mod             # Go module definition
├── finance.db         # SQLite database (created at runtime)
└── README.md          # This file
```

## 🏃 Getting Started

### Prerequisites
- Go 1.21 or higher
- SQLite3

### Installation

1. Navigate to project directory:
```bash
cd "Finance manager application backend "
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run server.go
```

The API will start on `http://localhost:8080`

## 📡 API Endpoints

### Authentication

#### Register User
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "message": "Registration successful",
  "user_id": 1
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "user_id": 1,
  "first_name": "John",
  "last_name": "Doe"
}
```

### Accounts

#### Get All Accounts
```http
GET /api/accounts?user_id=1
```

**Response:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "name": "Main Checking",
    "type": "checking",
    "currency": "USD",
    "initial_balance": 1000.00,
    "current_balance": 1500.50,
    "is_active": true,
    "created_at": "2026-01-09T10:00:00Z"
  }
]
```

#### Create Account
```http
POST /api/accounts
Content-Type: application/json

{
  "user_id": 1,
  "name": "Savings Account",
  "type": "savings",
  "currency": "USD",
  "initial_balance": 5000.00
}
```

### Transactions

#### Get Transactions
```http
GET /api/transactions?user_id=1&account_id=1
```

**Response:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "account_id": 1,
    "category_id": 1,
    "amount": 3000.00,
    "type": "credit",
    "date": "2026-01-09",
    "description": "Monthly Salary",
    "notes": "December payment",
    "created_at": "2026-01-09T10:00:00Z"
  }
]
```

#### Create Transaction
```http
POST /api/transactions
Content-Type: application/json

{
  "user_id": 1,
  "account_id": 1,
  "category_id": 4,
  "amount": 45.50,
  "type": "debit",
  "date": "2026-01-09",
  "description": "Grocery shopping",
  "notes": "Weekly groceries"
}
```

**Note:** Transaction types:
- `credit` - Money coming in (increases balance)
- `debit` - Money going out (decreases balance)

### Categories

#### Get All Categories
```http
GET /api/categories
```

**Response:**
```json
[
  {
    "id": 1,
    "user_id": null,
    "name": "Salary",
    "type": "income",
    "icon": "💼",
    "color": "#10b981"
  },
  {
    "id": 4,
    "user_id": null,
    "name": "Food & Dining",
    "type": "expense",
    "icon": "🍔",
    "color": "#ef4444"
  }
]
```

**Pre-loaded Categories:**

**Income:**
- Salary 💼
- Freelance 💻
- Investment 📈

**Expense:**
- Food & Dining 🍔
- Transportation 🚗
- Shopping 🛍️
- Entertainment 🎬
- Bills & Utilities 📄
- Healthcare 🏥
- Education 📚

### Budgets

#### Get User Budgets
```http
GET /api/budgets?user_id=1
```

**Response:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "category_id": 4,
    "limit_amount": 500.00,
    "period": "monthly",
    "start_date": "2026-01-01",
    "end_date": "2026-01-31",
    "created_at": "2026-01-09T10:00:00Z"
  }
]
```

#### Create Budget
```http
POST /api/budgets
Content-Type: application/json

{
  "user_id": 1,
  "category_id": 4,
  "limit_amount": 500.00,
  "period": "monthly",
  "start_date": "2026-01-01",
  "end_date": "2026-01-31"
}
```

**Budget Periods:** `daily`, `weekly`, `monthly`, `yearly`

## 🗄️ Database Schema

### Users Table
```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Accounts Table
```sql
CREATE TABLE accounts (
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
```

**Account Types:** `checking`, `savings`, `credit_card`, `cash`, `investment`

### Categories Table
```sql
CREATE TABLE categories (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  icon TEXT,
  color TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Category Types:** `income`, `expense`

### Transactions Table
```sql
CREATE TABLE transactions (
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
```

### Budgets Table
```sql
CREATE TABLE budgets (
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
);
```

## 🔐 Security Features

- **Password Hashing**: All passwords are hashed using bcrypt (cost factor: 10)
- **SQL Injection Prevention**: Parameterized queries throughout
- **CORS**: Configured for cross-origin requests
- **Unique Email Constraint**: Prevents duplicate user accounts

## 🧪 Testing with cURL

### Register a new user:
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123","first_name":"Test","last_name":"User"}'
```

### Login:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

### Create an account:
```bash
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"name":"Main Checking","type":"checking","currency":"USD","initial_balance":1000.00}'
```

### Get accounts:
```bash
curl http://localhost:8080/api/accounts?user_id=1
```

### Create a transaction:
```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"account_id":1,"category_id":1,"amount":3000.00,"type":"credit","date":"2026-01-09","description":"Monthly Salary"}'
```

### Get transactions:
```bash
curl http://localhost:8080/api/transactions?user_id=1
```

### Get categories:
```bash
curl http://localhost:8080/api/categories
```

### Create a budget:
```bash
curl -X POST http://localhost:8080/api/budgets \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"category_id":4,"limit_amount":500.00,"period":"monthly","start_date":"2026-01-01","end_date":"2026-01-31"}'
```

### Get budgets:
```bash
curl http://localhost:8080/api/budgets?user_id=1
```

## 📈 Roadmap

### ✅ Completed
- [x] User authentication (register, login)
- [x] Account management (create, list)
- [x] Transaction tracking (create, list, auto-balance update)
- [x] Category system (pre-loaded categories)
- [x] Budget management (create, list)
- [x] Database schema with foreign keys
- [x] Atomic transaction handling

### 🚧 Future Enhancements
- [ ] JWT token-based session management
- [ ] Update/Delete operations for accounts, transactions, budgets
- [ ] Transaction transfers between accounts
- [ ] Recurring transactions (subscriptions, bills)
- [ ] Financial reports & analytics
- [ ] Budget alerts and notifications
- [ ] Data export (CSV, PDF)
- [ ] Multi-currency conversion
- [ ] Transaction search and filters
- [ ] Dashboard statistics
- [ ] Frontend integration (React/Vue)
- [ ] Email notifications
- [ ] API rate limiting
- [ ] Comprehensive API documentation (Swagger)

## 🤝 Contributing

This is a portfolio project. Feel free to fork and extend!

## 📝 License

MIT License - feel free to use this for learning and portfolio purposes.

---

**Status**: 🟢 Running on http://localhost:8080
