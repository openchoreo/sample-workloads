package main

import (
	"encoding/gob"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type Poll struct {
	ID        string
	Question  string
	Options   []string
	Votes     map[string]int
	CreatedAt time.Time
}

type AppData struct {
	Polls map[string]*Poll
	mu    sync.RWMutex
}

var (
	appData = &AppData{
		Polls: make(map[string]*Poll),
	}
	store = sessions.NewCookieStore([]byte("super-secret-key-change-in-production"))
	tmpl  *template.Template
)

func init() {
	gob.Register(map[string]bool{})
}

func main() {
	var err error
	tmpl, err = template.New("").Funcs(template.FuncMap{
		"getPercentage": getPercentage,
		"getTotalVotes": getTotalVotes,
	}).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// Add some sample polls
	addSamplePolls()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/poll/", pollHandler)
	http.HandleFunc("/vote/", voteHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	appData.mu.RLock()
	defer appData.mu.RUnlock()

	// Sort polls by creation time (newest first)
	var polls []*Poll
	for _, poll := range appData.Polls {
		polls = append(polls, poll)
	}
	sort.Slice(polls, func(i, j int) bool {
		return polls[i].CreatedAt.After(polls[j].CreatedAt)
	})

	data := struct {
		Polls []*Poll
	}{
		Polls: polls,
	}

	if err := tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := tmpl.ExecuteTemplate(w, "create.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		question := r.FormValue("question")
		options := r.Form["options"]

		// Validate input
		if question == "" || len(options) < 2 {
			http.Error(w, "Question and at least 2 options are required", http.StatusBadRequest)
			return
		}

		// Filter out empty options
		var validOptions []string
		for _, opt := range options {
			if opt != "" {
				validOptions = append(validOptions, opt)
			}
		}

		if len(validOptions) < 2 {
			http.Error(w, "At least 2 valid options are required", http.StatusBadRequest)
			return
		}

		poll := &Poll{
			ID:        uuid.New().String(),
			Question:  question,
			Options:   validOptions,
			Votes:     make(map[string]int),
			CreatedAt: time.Now(),
		}

		// Initialize vote counts
		for _, opt := range validOptions {
			poll.Votes[opt] = 0
		}

		appData.mu.Lock()
		appData.Polls[poll.ID] = poll
		appData.mu.Unlock()

		http.Redirect(w, r, "/poll/"+poll.ID, http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func pollHandler(w http.ResponseWriter, r *http.Request) {
	pollID := r.URL.Path[len("/poll/"):]

	appData.mu.RLock()
	poll, exists := appData.Polls[pollID]
	appData.mu.RUnlock()

	if !exists {
		http.Error(w, "Poll not found", http.StatusNotFound)
		return
	}

	session, _ := store.Get(r, "voting-session")
	voted := false
	if votedPolls, ok := session.Values["voted"].(map[string]bool); ok {
		voted = votedPolls[pollID]
	}

	data := struct {
		Poll  *Poll
		Voted bool
	}{
		Poll:  poll,
		Voted: voted,
	}

	if err := tmpl.ExecuteTemplate(w, "poll.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func voteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pollID := r.URL.Path[len("/vote/"):]
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	option := r.FormValue("option")
	if option == "" {
		http.Error(w, "No option selected", http.StatusBadRequest)
		return
	}

	session, _ := store.Get(r, "voting-session")

	// Check if user has already voted
	votedPolls, ok := session.Values["voted"].(map[string]bool)
	if !ok {
		votedPolls = make(map[string]bool)
	}

	if votedPolls[pollID] {
		http.Redirect(w, r, "/poll/"+pollID, http.StatusSeeOther)
		return
	}

	appData.mu.Lock()
	poll, exists := appData.Polls[pollID]
	if exists {
		if _, validOption := poll.Votes[option]; validOption {
			poll.Votes[option]++
			votedPolls[pollID] = true
			session.Values["voted"] = votedPolls
			session.Save(r, w)
		}
	}
	appData.mu.Unlock()

	if !exists {
		http.Error(w, "Poll not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, "/poll/"+pollID, http.StatusSeeOther)
}

func getTotalVotes(poll *Poll) int {
	total := 0
	for _, count := range poll.Votes {
		total += count
	}
	return total
}

func getPercentage(votes int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(votes) / float64(total) * 100
}

func addSamplePolls() {
	samplePolls := []struct {
		question string
		options  []string
	}{
		{
			question: "What's your favorite programming language?",
			options:  []string{"Go", "Python", "JavaScript", "Java", "Rust"},
		},
		{
			question: "Preferred development environment?",
			options:  []string{"VS Code", "IntelliJ IDEA", "Vim/Neovim", "Sublime Text"},
		},
	}

	for _, sample := range samplePolls {
		poll := &Poll{
			ID:        uuid.New().String(),
			Question:  sample.question,
			Options:   sample.options,
			Votes:     make(map[string]int),
			CreatedAt: time.Now(),
		}

		for _, opt := range sample.options {
			poll.Votes[opt] = 0
		}

		appData.Polls[poll.ID] = poll
	}
}
