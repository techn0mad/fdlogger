# fdlogger

A self-contained, web-based app for Field Day logging using
Sqllite, Go, HTMX

```
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
```

## How to Run This

- Install Go: [https://go.dev/dl/](https://go.dev/dl/)

    In the project directory, run: `go mod init fdlogger`

- Install Go [sqllite3 module](https://github.com/mattn/go-sqlite3)

    `go get github.com/mattn/go-sqlite3` (or add to your `go.mod` if you use Go modules)

- Download htmx:

    Place [htmx.min.js](https://unpkg.com/htmx.org@1.9.10/dist/htmx.min.js) into your `static/` folder.

- Start the server:

    `go run main.go`

- Visit [http://localhost:8080](http://localhost:8080)


## How This Works

- Home page renders the log table and the entry form.
- Submitting the form sends an HTMX AJAX POST to /add, which inserts the row and responds with just the table HTML, swapped in-place.
- No need for any front-end Javascript other than [HTMX](https://htmx.org/) (which is a single file).
- You can add edit/delete, search, filtering, export, etc., just as simply.

## Why This is So Portable and Maintainable

- All dependencies are open source and easily vendored if needed.
- All logic, templates, and static assets are in your project directory.
- Upgrades: Just rebuild and redeploy the binary.
- Runs on Linux, Windows, Mac, Raspberry Pi: anywhere Go and SQLite work.

