package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
        "os"

	_ "github.com/mattn/go-sqlite3"
)

// LogEntry represents a contact log
type LogEntry struct {
	ID        int
	Callsign  string
	Time      string
	Frequency string
	Mode      string
	Notes     string
}

var (
	db        *sql.DB
	templates = template.Must(template.ParseGlob("templates/*.html"))
)

func main() {
	wd, _ := os.Getwd()
	log.Printf("Current working directory: %s", wd)

	var err error
	db, err = sql.Open("sqlite3", "./fieldday.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ensure table exists
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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/logs/row", rowHandler)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	entries := getAllLogs()
	err := templates.ExecuteTemplate(w, "base.html", entries)
        if err != nil {
            log.Printf("TEMPLATE ERROR: %v", err)
            http.Error(w, "Template error", 500)
        }
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "ParseForm error", 400)
		return
	}
	_, err := db.Exec(
		`INSERT INTO logs (callsign, time, frequency, mode, notes) VALUES (?, ?, ?, ?, ?)`,
		r.FormValue("callsign"), r.FormValue("time"), r.FormValue("frequency"), r.FormValue("mode"), r.FormValue("notes"),
	)
	if err != nil {
		http.Error(w, "DB insert error", 500)
		return
	}
	entries := getAllLogs()
	err = templates.ExecuteTemplate(w, "logs.html", entries)
        if err != nil {
            log.Printf("TEMPLATE ERROR (logs.html): %v", err)
            http.Error(w, "Template error", 500)
        }
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	row := db.QueryRow(`SELECT id, callsign, time, frequency, mode, notes FROM logs WHERE id = ?`, id)
	var e LogEntry
	if err := row.Scan(&e.ID, &e.Callsign, &e.Time, &e.Frequency, &e.Mode, &e.Notes); err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	err := templates.ExecuteTemplate(w, "edit_row.html", e)
        if err != nil { log.Printf("TEMPLATE ERROR (edit_row.html): %v", err)
            http.Error(w, "Template error", 500)
        }
}

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
	err = templates.ExecuteTemplate(w, "log_row.html", e)
        if err != nil {
            log.Printf("TEMPLATE ERROR (log_row.html): %v", err)
            http.Error(w, "Template error", 500)
        }
}

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

