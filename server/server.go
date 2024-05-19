package main

import (
	"database/sql"
	"log"
	"mailinglist/jsonapi"
	"mailinglist/mdb"
	"sync"

	"github.com/alexflint/go-arg"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

var args struct {
	DbPath   string `arg:"env:MAILINGLIST_DB"`
	BindJson string `arg:"env:MAILINGLIST_BIND_JSON"`
}

func main() {
	arg.MustParse(&args)

	// Set default values if not provided
	if args.DbPath == "" {
		args.DbPath = "list.db"
	}
	if args.BindJson == "" {
		args.BindJson = ":8080"
	}

	log.Printf("Server running in db path: %s\n", args.DbPath)
	db, err := sql.Open("sqlite3", args.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mdb.TryCreate(db)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done() // Ensure Done is called in a deferred manner
		log.Printf("Starting JSON API on %s\n", args.BindJson)
		jsonapi.Serve(db, args.BindJson)
	}()

	wg.Wait()
}
