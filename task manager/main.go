package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Priority int

const (
	Low Priority = iota
	Medium
	High
	Urgent
)

func (p Priority) String() string {
	switch p {
	case Low:
		return "LOW"
	case Medium:
		return "MEDIUM"
	case High:
		return "HIGH"
	case Urgent:
		return "URGENT"
	default:
		return "MEDIUM"
	}
}

func (p Priority) Color() string {
	switch p {
	case Low:
		return "\033[36m" // Cyan
	case Medium:
		return "\033[32m" // Green
	case High:
		return "\033[33m" // Yellow
	case Urgent:
		return "\033[31m" // Red
	default:
		return "\033[0m"
	}
}

func parsePriority(s string) Priority {
	switch strings.ToUpper(s) {
	case "LOW", "L", "1":
		return Low
	case "MEDIUM", "MED", "M", "2":
		return Medium
	case "HIGH", "H", "3":
		return High
	case "URGENT", "U", "4":
		return Urgent
	default:
		return Medium
	}
}

type TaskStatus string

const (
	StatusBacklog    TaskStatus = "backlog"
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusInReview   TaskStatus = "in_review"
	StatusDone       TaskStatus = "done"
)

type TimeEntry struct {
	ID        int        `json:"id"`
	TaskID    int        `json:"task_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Duration  int        `json:"duration"` // seconds
	Note      string     `json:"note,omitempty"`
}

type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Task struct {
	ID             int        `json:"id"`
	ProjectID      int        `json:"project_id"`
	Description    string     `json:"description"`
	Category       string     `json:"category,omitempty"`
	Priority       Priority   `json:"priority"`
	Status         TaskStatus `json:"status"`
	Done           bool       `json:"done"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
	Assignee       string     `json:"assignee,omitempty"`
	EstimatedHours float64    `json:"estimated_hours,omitempty"`
	Position       int        `json:"position"` // For kanban ordering
	Comments       []Comment  `json:"comments,omitempty"`
}

type Comment struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type AppData struct {
	Projects    []Project   `json:"projects"`
	Tasks       []Task      `json:"tasks"`
	TimeEntries []TimeEntry `json:"time_entries"`
}

func dataFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".project_manager.json"), nil
}

func loadAppData() (*AppData, error) {
	path, err := dataFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &AppData{
			Projects:    []Project{},
			Tasks:       []Task{},
			TimeEntries: []TimeEntry{},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	var appData AppData
	if err := json.Unmarshal(data, &appData); err != nil {
		return nil, err
	}
	return &appData, nil
}

func saveAppData(appData *AppData) error {
	path, err := dataFile()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(appData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func loadTasks() ([]Task, error) {
	appData, err := loadAppData()
	if err != nil {
		return nil, err
	}
	return appData.Tasks, nil
}

func saveTasks(tasks []Task) error {
	appData, err := loadAppData()
	if err != nil {
		return err
	}
	appData.Tasks = tasks
	return saveAppData(appData)
}

func nextID(tasks []Task) int {
	maxID := 0
	for _, t := range tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
}

func nextProjectID(projects []Project) int {
	maxID := 0
	for _, p := range projects {
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	return maxID + 1
}

func nextTimeEntryID(entries []TimeEntry) int {
	maxID := 0
	for _, e := range entries {
		if e.ID > maxID {
			maxID = e.ID
		}
	}
	return maxID + 1
}

func addTask(desc string) error {
	if strings.TrimSpace(desc) == "" {
		return errors.New("description cannot be empty")
	}
	appData, err := loadAppData()
	if err != nil {
		return err
	}

	// Get or create default project
	var defaultProject *Project
	for i := range appData.Projects {
		if appData.Projects[i].Name == "Default" {
			defaultProject = &appData.Projects[i]
			break
		}
	}
	if defaultProject == nil {
		newProject := Project{
			ID:        nextProjectID(appData.Projects),
			Name:      "Default",
			Color:     "#6366f1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		appData.Projects = append(appData.Projects, newProject)
		defaultProject = &newProject
	}

	task := Task{
		ID:          nextID(appData.Tasks),
		ProjectID:   defaultProject.ID,
		Description: desc,
		Priority:    Medium,
		Status:      StatusTodo,
		Done:        false,
		CreatedAt:   time.Now(),
		Position:    len(appData.Tasks),
	}
	appData.Tasks = append(appData.Tasks, task)
	if err := saveAppData(appData); err != nil {
		return err
	}
	fmt.Printf("✓ Added #%d: %s\n", task.ID, task.Description)
	return nil
}

func addTaskAdvanced(desc, category string, priority Priority, dueDate *time.Time, tags []string) error {
	if strings.TrimSpace(desc) == "" {
		return errors.New("description cannot be empty")
	}
	tasks, err := loadTasks()
	if err != nil {
		return err
	}
	task := Task{
		ID:          nextID(tasks),
		Description: desc,
		Category:    category,
		Priority:    priority,
		DueDate:     dueDate,
		Tags:        tags,
		Done:        false,
		CreatedAt:   time.Now(),
	}
	tasks = append(tasks, task)
	if err := saveTasks(tasks); err != nil {
		return err
	}
	fmt.Printf("✓ Added #%d: %s [%s%s\033[0m]\n", task.ID, task.Description, task.Priority.Color(), task.Priority)
	return nil
}

func listTasks() error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Println("No tasks yet.")
		return nil
	}

	// Sort by priority (highest first), then by due date
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Done != tasks[j].Done {
			return !tasks[i].Done // Incomplete tasks first
		}
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		if tasks[i].DueDate != nil && tasks[j].DueDate != nil {
			return tasks[i].DueDate.Before(*tasks[j].DueDate)
		}
		return tasks[i].ID < tasks[j].ID
	})

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("%-4s %-6s %-8s %-35s %-12s %s\n", "ID", "Status", "Priority", "Task", "Category", "Due Date")
	fmt.Println(strings.Repeat("=", 80))

	for _, t := range tasks {
		status := "[ ]"
		if t.Done {
			status = "[✓]"
		}

		dueStr := ""
		if t.DueDate != nil {
			dueStr = t.DueDate.Format("2006-01-02")
			if !t.Done && t.DueDate.Before(time.Now()) {
				dueStr = "\033[31m" + dueStr + " ⚠\033[0m"
			}
		}

		category := t.Category
		if category == "" {
			category = "general"
		}

		desc := t.Description
		if len(desc) > 35 {
			desc = desc[:32] + "..."
		}

		fmt.Printf("%-4d %s %s%-8s\033[0m %-35s %-12s %s",
			t.ID, status, t.Priority.Color(), t.Priority, desc, category, dueStr)

		if len(t.Tags) > 0 {
			fmt.Printf(" [%s]", strings.Join(t.Tags, ", "))
		}
		fmt.Println()
	}
	fmt.Println(strings.Repeat("=", 80))
	return nil
}

func listByCategory(category string) error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	filtered := []Task{}
	for _, t := range tasks {
		if strings.EqualFold(t.Category, category) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("No tasks in category '%s'\n", category)
		return nil
	}

	fmt.Printf("\nTasks in category '%s':\n", category)
	for _, t := range filtered {
		status := " "
		if t.Done {
			status = "✓"
		}
		fmt.Printf("[%s] #%d %s%s\033[0m: %s\n", status, t.ID, t.Priority.Color(), t.Priority, t.Description)
	}
	return nil
}

func searchTasks(query string) error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	query = strings.ToLower(query)
	found := []Task{}

	for _, t := range tasks {
		if strings.Contains(strings.ToLower(t.Description), query) ||
			strings.Contains(strings.ToLower(t.Category), query) {
			found = append(found, t)
		} else {
			for _, tag := range t.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					found = append(found, t)
					break
				}
			}
		}
	}

	if len(found) == 0 {
		fmt.Printf("No tasks found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("\nFound %d task(s) matching '%s':\n", len(found), query)
	for _, t := range found {
		status := " "
		if t.Done {
			status = "✓"
		}
		fmt.Printf("[%s] #%d %s%s\033[0m: %s\n", status, t.ID, t.Priority.Color(), t.Priority, t.Description)
	}
	return nil
}

func viewTask(id int) error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	for _, t := range tasks {
		if t.ID == id {
			fmt.Println("\n" + strings.Repeat("=", 60))
			fmt.Printf("Task #%d\n", t.ID)
			fmt.Println(strings.Repeat("=", 60))
			fmt.Printf("Description:  %s\n", t.Description)
			fmt.Printf("Status:       %s\n", map[bool]string{true: "✓ Completed", false: "○ Pending"}[t.Done])
			fmt.Printf("Priority:     %s%s\033[0m\n", t.Priority.Color(), t.Priority)
			if t.Category != "" {
				fmt.Printf("Category:     %s\n", t.Category)
			}
			if t.DueDate != nil {
				fmt.Printf("Due Date:     %s\n", t.DueDate.Format("2006-01-02 15:04"))
				if !t.Done && t.DueDate.Before(time.Now()) {
					fmt.Printf("              \033[31m⚠ OVERDUE\033[0m\n")
				}
			}
			fmt.Printf("Created:      %s\n", t.CreatedAt.Format("2006-01-02 15:04"))
			if t.CompletedAt != nil {
				fmt.Printf("Completed:    %s\n", t.CompletedAt.Format("2006-01-02 15:04"))
			}
			if len(t.Tags) > 0 {
				fmt.Printf("Tags:         %s\n", strings.Join(t.Tags, ", "))
			}
			fmt.Println(strings.Repeat("=", 60))
			return nil
		}
	}
	return fmt.Errorf("task #%d not found", id)
}

func updatePriority(id int, priority Priority) error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Priority = priority
			if err := saveTasks(tasks); err != nil {
				return err
			}
			fmt.Printf("✓ Updated #%d priority to %s%s\033[0m\n", id, priority.Color(), priority)
			return nil
		}
	}
	return fmt.Errorf("task #%d not found", id)
}

func setDueDate(id int, dueDate time.Time) error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].DueDate = &dueDate
			if err := saveTasks(tasks); err != nil {
				return err
			}
			fmt.Printf("✓ Set due date for #%d to %s\n", id, dueDate.Format("2006-01-02"))
			return nil
		}
	}
	return fmt.Errorf("task #%d not found", id)
}

func showStats() error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	total := len(tasks)
	completed := 0
	overdue := 0
	byPriority := map[Priority]int{Low: 0, Medium: 0, High: 0, Urgent: 0}
	categories := make(map[string]int)

	for _, t := range tasks {
		if t.Done {
			completed++
		} else {
			byPriority[t.Priority]++
			if t.DueDate != nil && t.DueDate.Before(time.Now()) {
				overdue++
			}
		}
		if t.Category != "" {
			categories[t.Category]++
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📊 Task Statistics")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total Tasks:      %d\n", total)
	fmt.Printf("Completed:        %d (%.1f%%)\n", completed, float64(completed)/float64(max(total, 1))*100)
	fmt.Printf("Pending:          %d\n", total-completed)
	fmt.Printf("Overdue:          %d\n", overdue)
	fmt.Println("\nBy Priority (pending only):")
	fmt.Printf("  Urgent:         %d\n", byPriority[Urgent])
	fmt.Printf("  High:           %d\n", byPriority[High])
	fmt.Printf("  Medium:         %d\n", byPriority[Medium])
	fmt.Printf("  Low:            %d\n", byPriority[Low])

	if len(categories) > 0 {
		fmt.Println("\nBy Category:")
		for cat, count := range categories {
			fmt.Printf("  %-15s %d\n", cat+":", count)
		}
	}
	fmt.Println(strings.Repeat("=", 50))
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func markDone(id int) error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}
	found := false
	now := time.Now()
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Done = true
			tasks[i].CompletedAt = &now
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("task #%d not found", id)
	}
	if err := saveTasks(tasks); err != nil {
		return err
	}
	fmt.Printf("✓ Completed #%d\n", id)
	return nil
}

func deleteTask(id int) error {
	tasks, err := loadTasks()
	if err != nil {
		return err
	}
	out := tasks[:0]
	found := false
	for _, t := range tasks {
		if t.ID == id {
			found = true
			continue
		}
		out = append(out, t)
	}
	if !found {
		return fmt.Errorf("task #%d not found", id)
	}
	if err := saveTasks(out); err != nil {
		return err
	}
	fmt.Printf("✓ Deleted #%d\n", id)
	return nil
}

func parseDate(dateStr string) (*time.Time, error) {
	// Try various date formats
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01-02-2006",
		"01/02/2006",
		"Jan 02 2006",
		"02 Jan 2006",
	}

	// Handle relative dates
	lower := strings.ToLower(dateStr)
	now := time.Now()

	switch lower {
	case "today":
		t := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		return &t, nil
	case "tomorrow":
		t := now.AddDate(0, 0, 1)
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		return &t, nil
	}

	// Check if it's "in X days"
	if strings.HasPrefix(lower, "in ") && strings.HasSuffix(lower, " days") {
		daysStr := strings.TrimPrefix(lower, "in ")
		daysStr = strings.TrimSuffix(daysStr, " days")
		days, err := strconv.Atoi(daysStr)
		if err == nil {
			t := now.AddDate(0, 0, days)
			t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
			return &t, nil
		}
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
			return &t, nil
		}
	}

	return nil, fmt.Errorf("invalid date format. Try: YYYY-MM-DD, today, tomorrow, or 'in X days'")
}

func usage() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📋 CLI Task Manager - Available Commands")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\nBasic Commands:")
	fmt.Println("  add <description>                    - Add a simple task")
	fmt.Println("  list                                 - List all tasks")
	fmt.Println("  done <id>                            - Mark task as complete")
	fmt.Println("  delete <id>                          - Delete a task")
	fmt.Println("  view <id>                            - View task details")
	fmt.Println("\nAdvanced Commands:")
	fmt.Println("  create --desc \"...\" [options]        - Create task with options")
	fmt.Println("    Options:")
	fmt.Println("      --priority <low|medium|high|urgent>")
	fmt.Println("      --category <name>")
	fmt.Println("      --due <date>                     - Date format: YYYY-MM-DD, today, tomorrow")
	fmt.Println("      --tags <tag1,tag2>")
	fmt.Println("\nFilter & Search:")
	fmt.Println("  search <query>                       - Search tasks by keyword")
	fmt.Println("  category <name>                      - List tasks by category")
	fmt.Println("  stats                                - Show task statistics")
	fmt.Println("\nUpdate Commands:")
	fmt.Println("  priority <id> <low|medium|high|urgent> - Update task priority")
	fmt.Println("  due <id> <date>                      - Set/update due date")
	fmt.Println("\nOther:")
	fmt.Println("  help                                 - Show this help")
	fmt.Println("  clear                                - Clear screen")
	fmt.Println("  quit/exit                            - Exit program")
	fmt.Println(strings.Repeat("=", 70))
}

func parseCreateCommand(args []string) (desc, category string, priority Priority, dueDate *time.Time, tags []string, err error) {
	priority = Medium

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--desc", "-d":
			if i+1 >= len(args) {
				err = errors.New("--desc requires a value")
				return
			}
			desc = args[i+1]
			i++
		case "--priority", "-p":
			if i+1 >= len(args) {
				err = errors.New("--priority requires a value")
				return
			}
			priority = parsePriority(args[i+1])
			i++
		case "--category", "-c":
			if i+1 >= len(args) {
				err = errors.New("--category requires a value")
				return
			}
			category = args[i+1]
			i++
		case "--due":
			if i+1 >= len(args) {
				err = errors.New("--due requires a value")
				return
			}
			dueDate, err = parseDate(args[i+1])
			if err != nil {
				return
			}
			i++
		case "--tags", "-t":
			if i+1 >= len(args) {
				err = errors.New("--tags requires a value")
				return
			}
			tags = strings.Split(args[i+1], ",")
			for j := range tags {
				tags[j] = strings.TrimSpace(tags[j])
			}
			i++
		}
	}

	if desc == "" {
		err = errors.New("description is required (use --desc)")
	}
	return
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	// Check if server mode is requested
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startServer()
		return
	}

	clearScreen()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           📋 CLI Task Manager - Interactive Mode            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println("\nType 'help' for available commands or 'quit' to exit")
	fmt.Println("💡 Tip: Run 'go run . server' to start the web interface\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n\033[36m>\033[0m ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := parts[0]

		switch cmd {
		case "add":
			if len(parts) < 2 {
				fmt.Println("Error: add requires a description")
				continue
			}
			desc := strings.Join(parts[1:], " ")
			if err := addTask(desc); err != nil {
				fmt.Println("Error:", err)
			}

		case "create":
			desc, category, priority, dueDate, tags, err := parseCreateCommand(parts[1:])
			if err != nil {
				fmt.Println("Error:", err)
				fmt.Println("Usage: create --desc \"task description\" [--priority low|medium|high|urgent] [--category name] [--due YYYY-MM-DD] [--tags tag1,tag2]")
				continue
			}
			if err := addTaskAdvanced(desc, category, priority, dueDate, tags); err != nil {
				fmt.Println("Error:", err)
			}

		case "list":
			if err := listTasks(); err != nil {
				fmt.Println("Error:", err)
			}

		case "view":
			if len(parts) < 2 {
				fmt.Println("Error: view requires an id")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Error: id must be a number")
				continue
			}
			if err := viewTask(id); err != nil {
				fmt.Println("Error:", err)
			}

		case "done":
			if len(parts) < 2 {
				fmt.Println("Error: done requires an id")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Error: id must be a number")
				continue
			}
			if err := markDone(id); err != nil {
				fmt.Println("Error:", err)
			}

		case "delete", "del":
			if len(parts) < 2 {
				fmt.Println("Error: delete requires an id")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Error: id must be a number")
				continue
			}
			if err := deleteTask(id); err != nil {
				fmt.Println("Error:", err)
			}

		case "priority":
			if len(parts) < 3 {
				fmt.Println("Error: priority requires an id and priority level")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Error: id must be a number")
				continue
			}
			priority := parsePriority(parts[2])
			if err := updatePriority(id, priority); err != nil {
				fmt.Println("Error:", err)
			}

		case "due":
			if len(parts) < 3 {
				fmt.Println("Error: due requires an id and date")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Error: id must be a number")
				continue
			}
			dueDate, err := parseDate(parts[2])
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			if err := setDueDate(id, *dueDate); err != nil {
				fmt.Println("Error:", err)
			}

		case "search":
			if len(parts) < 2 {
				fmt.Println("Error: search requires a query")
				continue
			}
			query := strings.Join(parts[1:], " ")
			if err := searchTasks(query); err != nil {
				fmt.Println("Error:", err)
			}

		case "category", "cat":
			if len(parts) < 2 {
				fmt.Println("Error: category requires a category name")
				continue
			}
			if err := listByCategory(parts[1]); err != nil {
				fmt.Println("Error:", err)
			}

		case "stats":
			if err := showStats(); err != nil {
				fmt.Println("Error:", err)
			}

		case "clear", "cls":
			clearScreen()

		case "help", "h", "?":
			usage()

		case "quit", "exit", "q":
			fmt.Println("\n👋 Goodbye! Stay productive!")
			return

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}
