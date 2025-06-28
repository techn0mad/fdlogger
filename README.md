# fdlogger

A self-contained, web-based app for Field Day logging

.
├── README.md
├── main.go
├── static
│   └── htmx.min.js
└── templates
    ├── base.html
    ├── edit_row.html
    ├── index.html
    ├── log_row.html
    └── logs.html

3 directories, 8 files

## How to Run This

Install Go: https://go.dev/dl/

Install dependency:
    In the project directory, run:
    go mod init fdlogger
    go get github.com/mattn/go-sqlite3

Download htmx:
    Place htmx.min.js in your static/ folder.

Start the server:
    go run main.go

Visit:
    http://localhost:8080
