package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Student represents a student record
type Student struct {
	ID             int       `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	DateOfBirth    string    `json:"date_of_birth"`
	Address        string    `json:"address"`
	EnrollmentDate string    `json:"enrollment_date"`
	Status         string    `json:"status"` // active, inactive, graduated
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Course represents a course
type Course struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Credits     int       `json:"credits"`
	Instructor  string    `json:"instructor"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Enrollment represents student course enrollment
type Enrollment struct {
	ID             int       `json:"id"`
	StudentID      int       `json:"student_id"`
	CourseID       int       `json:"course_id"`
	EnrollmentDate string    `json:"enrollment_date"`
	Grade          string    `json:"grade,omitempty"`
	Status         string    `json:"status"` // enrolled, completed, dropped
	CreatedAt      time.Time `json:"created_at"`
}

// Response wrapper
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./students.db")
	if err != nil {
		return err
	}

	// Create tables
	createTables := `
	CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT,
		date_of_birth TEXT,
		address TEXT,
		enrollment_date TEXT,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS courses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		credits INTEGER DEFAULT 3,
		instructor TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS enrollments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER NOT NULL,
		course_id INTEGER NOT NULL,
		enrollment_date TEXT,
		grade TEXT,
		status TEXT DEFAULT 'enrolled',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
		FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
		UNIQUE(student_id, course_id)
	);
	`

	_, err = db.Exec(createTables)
	return err
}

// CORS middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Student handlers
func getStudents(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, first_name, last_name, email, phone, date_of_birth, 
		       address, enrollment_date, status, created_at, updated_at 
		FROM students ORDER BY created_at DESC
	`)
	if err != nil {
		sendJSON(w, Response{Success: false, Message: err.Error()})
		return
	}
	defer rows.Close()

	students := []Student{}
	for rows.Next() {
		var s Student
		err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.Phone,
			&s.DateOfBirth, &s.Address, &s.EnrollmentDate, &s.Status,
			&s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			continue
		}
		students = append(students, s)
	}

	sendJSON(w, Response{Success: true, Data: students})
}

func getStudent(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid ID"})
		return
	}

	var s Student
	err = db.QueryRow(`
		SELECT id, first_name, last_name, email, phone, date_of_birth,
		       address, enrollment_date, status, created_at, updated_at
		FROM students WHERE id = ?
	`, id).Scan(&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.Phone,
		&s.DateOfBirth, &s.Address, &s.EnrollmentDate, &s.Status,
		&s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Student not found"})
		return
	}

	sendJSON(w, Response{Success: true, Data: s})
}

func createStudent(w http.ResponseWriter, r *http.Request) {
	var s Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid request"})
		return
	}

	result, err := db.Exec(`
		INSERT INTO students (first_name, last_name, email, phone, date_of_birth, 
		                     address, enrollment_date, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, s.FirstName, s.LastName, s.Email, s.Phone, s.DateOfBirth,
		s.Address, s.EnrollmentDate, s.Status)

	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Failed to create student: " + err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	s.ID = int(id)
	sendJSON(w, Response{Success: true, Message: "Student created successfully", Data: s})
}

func updateStudent(w http.ResponseWriter, r *http.Request) {
	var s Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid request"})
		return
	}

	_, err := db.Exec(`
		UPDATE students 
		SET first_name = ?, last_name = ?, email = ?, phone = ?, 
		    date_of_birth = ?, address = ?, enrollment_date = ?, 
		    status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, s.FirstName, s.LastName, s.Email, s.Phone, s.DateOfBirth,
		s.Address, s.EnrollmentDate, s.Status, s.ID)

	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Failed to update student"})
		return
	}

	sendJSON(w, Response{Success: true, Message: "Student updated successfully", Data: s})
}

func deleteStudent(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid ID"})
		return
	}

	_, err = db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Failed to delete student"})
		return
	}

	sendJSON(w, Response{Success: true, Message: "Student deleted successfully"})
}

// Course handlers
func getCourses(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, code, name, description, credits, instructor, created_at, updated_at
		FROM courses ORDER BY code
	`)
	if err != nil {
		sendJSON(w, Response{Success: false, Message: err.Error()})
		return
	}
	defer rows.Close()

	courses := []Course{}
	for rows.Next() {
		var c Course
		err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Credits,
			&c.Instructor, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			continue
		}
		courses = append(courses, c)
	}

	sendJSON(w, Response{Success: true, Data: courses})
}

func createCourse(w http.ResponseWriter, r *http.Request) {
	var c Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid request"})
		return
	}

	result, err := db.Exec(`
		INSERT INTO courses (code, name, description, credits, instructor)
		VALUES (?, ?, ?, ?, ?)
	`, c.Code, c.Name, c.Description, c.Credits, c.Instructor)

	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Failed to create course"})
		return
	}

	id, _ := result.LastInsertId()
	c.ID = int(id)
	sendJSON(w, Response{Success: true, Message: "Course created successfully", Data: c})
}

func updateCourse(w http.ResponseWriter, r *http.Request) {
	var c Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid request"})
		return
	}

	_, err := db.Exec(`
		UPDATE courses 
		SET code = ?, name = ?, description = ?, credits = ?, 
		    instructor = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, c.Code, c.Name, c.Description, c.Credits, c.Instructor, c.ID)

	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Failed to update course"})
		return
	}

	sendJSON(w, Response{Success: true, Message: "Course updated successfully", Data: c})
}

func deleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid ID"})
		return
	}

	_, err = db.Exec("DELETE FROM courses WHERE id = ?", id)
	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Failed to delete course"})
		return
	}

	sendJSON(w, Response{Success: true, Message: "Course deleted successfully"})
}

// Enrollment handlers
func getEnrollments(w http.ResponseWriter, r *http.Request) {
	studentID := r.URL.Query().Get("student_id")
	courseID := r.URL.Query().Get("course_id")

	query := `
		SELECT e.id, e.student_id, e.course_id, e.enrollment_date, 
		       e.grade, e.status, e.created_at
		FROM enrollments e
		WHERE 1=1
	`
	args := []interface{}{}

	if studentID != "" {
		query += " AND e.student_id = ?"
		args = append(args, studentID)
	}

	if courseID != "" {
		query += " AND e.course_id = ?"
		args = append(args, courseID)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		sendJSON(w, Response{Success: false, Message: err.Error()})
		return
	}
	defer rows.Close()

	enrollments := []Enrollment{}
	for rows.Next() {
		var e Enrollment
		err := rows.Scan(&e.ID, &e.StudentID, &e.CourseID, &e.EnrollmentDate,
			&e.Grade, &e.Status, &e.CreatedAt)
		if err != nil {
			continue
		}
		enrollments = append(enrollments, e)
	}

	sendJSON(w, Response{Success: true, Data: enrollments})
}

func createEnrollment(w http.ResponseWriter, r *http.Request) {
	var e Enrollment
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid request"})
		return
	}

	result, err := db.Exec(`
		INSERT INTO enrollments (student_id, course_id, enrollment_date, status)
		VALUES (?, ?, ?, ?)
	`, e.StudentID, e.CourseID, e.EnrollmentDate, e.Status)

	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Failed to enroll student: " + err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	e.ID = int(id)
	sendJSON(w, Response{Success: true, Message: "Enrollment successful", Data: e})
}

func updateEnrollment(w http.ResponseWriter, r *http.Request) {
	var e Enrollment
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid request"})
		return
	}

	_, err := db.Exec(`
		UPDATE enrollments 
		SET grade = ?, status = ?
		WHERE id = ?
	`, e.Grade, e.Status, e.ID)

	if err != nil {
		sendJSON(w, Response{Success: false, Message: "Failed to update enrollment"})
		return
	}

	sendJSON(w, Response{Success: true, Message: "Enrollment updated successfully", Data: e})
}

func getDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})

	// Total students
	var totalStudents, activeStudents int
	db.QueryRow("SELECT COUNT(*) FROM students").Scan(&totalStudents)
	db.QueryRow("SELECT COUNT(*) FROM students WHERE status = 'active'").Scan(&activeStudents)

	// Total courses
	var totalCourses int
	db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&totalCourses)

	// Total enrollments
	var totalEnrollments int
	db.QueryRow("SELECT COUNT(*) FROM enrollments WHERE status = 'enrolled'").Scan(&totalEnrollments)

	stats["total_students"] = totalStudents
	stats["active_students"] = activeStudents
	stats["total_courses"] = totalCourses
	stats["total_enrollments"] = totalEnrollments

	sendJSON(w, Response{Success: true, Data: stats})
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Serve static files
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	// API routes
	http.HandleFunc("/api/students", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Query().Get("id") != "" {
				getStudent(w, r)
			} else {
				getStudents(w, r)
			}
		case "POST":
			createStudent(w, r)
		case "PUT":
			updateStudent(w, r)
		case "DELETE":
			deleteStudent(w, r)
		}
	}))

	http.HandleFunc("/api/courses", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getCourses(w, r)
		case "POST":
			createCourse(w, r)
		case "PUT":
			updateCourse(w, r)
		case "DELETE":
			deleteCourse(w, r)
		}
	}))

	http.HandleFunc("/api/enrollments", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getEnrollments(w, r)
		case "POST":
			createEnrollment(w, r)
		case "PUT":
			updateEnrollment(w, r)
		}
	}))

	http.HandleFunc("/api/stats", enableCORS(getDashboardStats))

	fmt.Println("🎓 Student Management System started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
