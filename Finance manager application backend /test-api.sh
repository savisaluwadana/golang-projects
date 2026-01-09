#!/bin/bash

# Finance Manager API Test Script
# This script tests all endpoints of the Finance Manager API

BASE_URL="http://localhost:8080"

echo "🧪 Testing Finance Manager API"
echo "================================"
echo ""

# Test 1: Register User
echo "1. Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@test.com","password":"demo123","first_name":"Demo","last_name":"User"}')
echo "✅ Register Response: $REGISTER_RESPONSE"
USER_ID=$(echo $REGISTER_RESPONSE | grep -o '"user_id":[0-9]*' | grep -o '[0-9]*')
echo "   User ID: $USER_ID"
echo ""

# Test 2: Login
echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@test.com","password":"demo123"}')
echo "✅ Login Response: $LOGIN_RESPONSE"
echo ""

# Test 3: Create Account
echo "3. Creating checking account..."
ACCOUNT_RESPONSE=$(curl -s -X POST $BASE_URL/api/accounts \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":$USER_ID,\"name\":\"Main Checking\",\"type\":\"checking\",\"currency\":\"USD\",\"initial_balance\":1000.00}")
echo "✅ Account Response: $ACCOUNT_RESPONSE"
ACCOUNT_ID=$(echo $ACCOUNT_RESPONSE | grep -o '"id":[0-9]*' | head -1 | grep -o '[0-9]*')
echo "   Account ID: $ACCOUNT_ID"
echo ""

# Test 4: Get Accounts
echo "4. Fetching all accounts..."
ACCOUNTS=$(curl -s "$BASE_URL/api/accounts?user_id=$USER_ID")
echo "✅ Accounts: $ACCOUNTS"
echo ""

# Test 5: Get Categories
echo "5. Fetching categories..."
CATEGORIES=$(curl -s "$BASE_URL/api/categories")
echo "✅ Categories: $CATEGORIES"
echo ""

# Test 6: Create Income Transaction
echo "6. Creating income transaction (Salary)..."
INCOME_TX=$(curl -s -X POST $BASE_URL/api/transactions \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":$USER_ID,\"account_id\":$ACCOUNT_ID,\"category_id\":1,\"amount\":3000.00,\"type\":\"credit\",\"date\":\"2026-01-09\",\"description\":\"Monthly Salary\",\"notes\":\"January payment\"}")
echo "✅ Income Transaction: $INCOME_TX"
echo ""

# Test 7: Create Expense Transaction
echo "7. Creating expense transaction (Food)..."
EXPENSE_TX=$(curl -s -X POST $BASE_URL/api/transactions \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":$USER_ID,\"account_id\":$ACCOUNT_ID,\"category_id\":4,\"amount\":45.50,\"type\":\"debit\",\"date\":\"2026-01-09\",\"description\":\"Grocery Shopping\",\"notes\":\"Weekly groceries\"}")
echo "✅ Expense Transaction: $EXPENSE_TX"
echo ""

# Test 8: Get Transactions
echo "8. Fetching all transactions..."
TRANSACTIONS=$(curl -s "$BASE_URL/api/transactions?user_id=$USER_ID")
echo "✅ Transactions: $TRANSACTIONS"
echo ""

# Test 9: Get Updated Account Balance
echo "9. Checking updated account balance..."
UPDATED_ACCOUNTS=$(curl -s "$BASE_URL/api/accounts?user_id=$USER_ID")
echo "✅ Updated Accounts: $UPDATED_ACCOUNTS"
echo "   Expected Balance: 1000 + 3000 - 45.50 = 3954.50"
echo ""

# Test 10: Create Budget
echo "10. Creating monthly budget for Food..."
BUDGET=$(curl -s -X POST $BASE_URL/api/budgets \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":$USER_ID,\"category_id\":4,\"limit_amount\":500.00,\"period\":\"monthly\",\"start_date\":\"2026-01-01\",\"end_date\":\"2026-01-31\"}")
echo "✅ Budget: $BUDGET"
echo ""

# Test 11: Get Budgets
echo "11. Fetching all budgets..."
BUDGETS=$(curl -s "$BASE_URL/api/budgets?user_id=$USER_ID")
echo "✅ Budgets: $BUDGETS"
echo ""

echo "================================"
echo "✨ All tests completed!"
echo ""
echo "Summary:"
echo "- User registered and logged in"
echo "- Account created with \$1000 initial balance"
echo "- Income transaction: +\$3000"
echo "- Expense transaction: -\$45.50"
echo "- Final balance should be: \$3954.50"
echo "- Budget created for Food category: \$500/month"
