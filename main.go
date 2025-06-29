package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3" // Registers the SQLite3 driver for sql.Open()
)

// ----------------------------
// Data structures
// ----------------------------

// LogEntry represents a single log record (row in DB).
type LogEntry struct {
	ID        int
	Callsign  string
	Time      string
	Frequency string
	Mode      string
	Notes     string
}

// List of allowed radio modes (used for dropdowns in forms).
var allowedModes = []string{"CW", "SSB", "FM", "Digital"}

// PageData is used for templates rendering the main page or entry area.
type PageData struct {
	Entries []LogEntry
	Modes   []string
}

// EditRowData is for rendering a single log entry with the mode list, e.g. in the edit form.
type EditRowData struct {
	LogEntry
	Modes []string
}

// ----------------------------
// Global variables
// ----------------------------

var (
	db        *sql.DB                                                 // Global DB connection (server-side)
	templates = template.Must(template.ParseGlob("templates/*.html")) // Parses all templates at startup (server-side)
)

// ----------------------------
// main() - Server setup
// ----------------------------

func main() {
	wd, _ := os.Getwd()
	log.Printf("Current working directory: %s", wd)

	var err error
	// Opens (or creates) the SQLite DB file in the project directory (server-side)
	db, err = sql.Open("sqlite3", "./fieldday.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Make sure the "logs" table exists
	createTable := `
    CREATE TABLE IF NOT EXISTS logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        callsign TEXT,
        time TEXT,
        frequency TEXT,
        mode TEXT,
        notes TEXT
    );`
	if _, err = db.Exec(createTable); err != nil {
		log.Fatal(err)
	}

	// Routing: maps URLs to handler functions (server-side)
	// The explicit connection from URL path to handler is made http.HandleFunc().

	// Serves files (like htmx.min.js) from the static/ directory when a browser requests /static/htmx.min.js.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static")))) // Serves JS/CSS files

	// Each line below registers a handler function to a specific URL path.
	// When the Go HTTP server receives a request to a URL, it calls the matching function.
	http.HandleFunc("/", indexHandler)        // Main page
	http.HandleFunc("/add", addHandler)       // Add new entry (POST, via HTMX)
	http.HandleFunc("/edit", editHandler)     // Edit form for a row (GET, via HTMX)
	http.HandleFunc("/update", updateHandler) // Save changes to a row (POST, via HTMX)
	http.HandleFunc("/logs/row", rowHandler)  // Rerender a single row (e.g. on cancel, via HTMX)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil)) // Starts HTTP server (server-side)
}

// ----------------------------
// Handler: Main page
// ----------------------------

// Handles GET "/" - renders the main page with the entry form and log table
func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Entries: getAllLogs(), // Fetch all entries from the DB (server-side)
		Modes:   allowedModes, // Supply mode options for the form
	}
	// Render full page using base.html layout, with .Entries and .Modes
	err := templates.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("TEMPLATE ERROR (base.html): %v", err)
		http.Error(w, "Template error", 500)
	}
}

// ----------------------------
// Handler: Add a new entry
// ----------------------------

// Handles POST "/add" - inserts a new log entry, returns a fragment (form + table)
// This is called by the browser when the user submits the Add Log form
func addHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "ParseForm error", 400)
		return
	}
	// Insert the new record into the DB
	_, err := db.Exec(
		`INSERT INTO logs (callsign, time, frequency, mode, notes) VALUES (?, ?, ?, ?, ?)`,
		r.FormValue("callsign"), r.FormValue("time"), r.FormValue("frequency"), r.FormValue("mode"), r.FormValue("notes"),
	)
	if err != nil {
		http.Error(w, "DB insert error", 500)
		return
	}
	data := PageData{
		Entries: getAllLogs(), // Refresh entries for updated table
		Modes:   allowedModes, // Mode options for the (reset) form
	}
	// Render entryarea.html: a fragment containing a fresh form and table (swapped by HTMX)
	err = templates.ExecuteTemplate(w, "entryarea.html", data)
	if err != nil {
		log.Printf("TEMPLATE ERROR (entryarea.html): %v", err)
		http.Error(w, "Template error", 500)
	}
}

// ----------------------------
// Handler: Show edit form for a row
// ----------------------------

// Handles GET "/edit?id=NN" - returns the edit form fragment for the specified row
// (HTMX swaps this into the table in the browser)
func editHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	row := db.QueryRow(`SELECT id, callsign, time, frequency, mode, notes FROM logs WHERE id = ?`, id)
	var e LogEntry
	if err := row.Scan(&e.ID, &e.Callsign, &e.Time, &e.Frequency, &e.Mode, &e.Notes); err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	data := EditRowData{
		LogEntry: e,
		Modes:    allowedModes, // Pass in all possible modes for the select list
	}
	// Render edit_row.html fragment (a single row replaced by HTMX in the browser)
	err := templates.ExecuteTemplate(w, "edit_row.html", data)
	if err != nil {
		log.Printf("TEMPLATE ERROR (edit_row.html): %v", err)
		http.Error(w, "Template error", 500)
	}
}

// ----------------------------
// Handler: Save an edit (update a row)
// ----------------------------

// Handles POST "/update" - updates a record in the DB, returns the updated row fragment
func updateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "ParseForm error", 400)
		return
	}
	id := r.FormValue("id")
	_, err := db.Exec(
		`UPDATE logs SET callsign=?, time=?, frequency=?, mode=?, notes=? WHERE id=?`,
		r.FormValue("callsign"), r.FormValue("time"), r.FormValue("frequency"),
		r.FormValue("mode"), r.FormValue("notes"), id,
	)
	if err != nil {
		http.Error(w, "DB update error", 500)
		return
	}
	row := db.QueryRow(`SELECT id, callsign, time, frequency, mode, notes FROM logs WHERE id = ?`, id)
	var e LogEntry
	row.Scan(&e.ID, &e.Callsign, &e.Time, &e.Frequency, &e.Mode, &e.Notes)
	// Render just the updated row as an HTML fragment
	err = templates.ExecuteTemplate(w, "log_row.html", e)
	if err != nil {
		log.Printf("TEMPLATE ERROR (log_row.html): %v", err)
		http.Error(w, "Template error", 500)
	}
}

// ----------------------------
// Handler: Show a row (used for Cancel)
// ----------------------------

// Handles GET "/logs/row?id=NN" - returns just the table row HTML fragment for the given log entry
func rowHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	row := db.QueryRow(`SELECT id, callsign, time, frequency, mode, notes FROM logs WHERE id = ?`, id)
	var e LogEntry
	if err := row.Scan(&e.ID, &e.Callsign, &e.Time, &e.Frequency, &e.Mode, &e.Notes); err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	err := templates.ExecuteTemplate(w, "log_row.html", e)
	if err != nil {
		log.Printf("TEMPLATE ERROR (log_row.html): %v", err)
		http.Error(w, "Template error", 500)
	}
}

// ----------------------------
// Helper: Get all logs from the DB
// ----------------------------

// Fetches all log entries as a slice (server-side only, never exposed directly to browser)
func getAllLogs() []LogEntry {
	rows, err := db.Query(`SELECT id, callsign, time, frequency, mode, notes FROM logs ORDER BY id DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		rows.Scan(&e.ID, &e.Callsign, &e.Time, &e.Frequency, &e.Mode, &e.Notes)
		entries = append(entries, e)
	}
	return entries
}
