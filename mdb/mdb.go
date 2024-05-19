package mdb

import (
	"database/sql"
	"log"
	"time"
)

// 	"github.com/mattn/go-sqlite3"

type EmailEntry struct {
	Id          int
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
}

func TryCreate(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS emails (id INTEGER PRIMARY KEY, email TEXT UNIQUE, confirmed_at INTEGER, opt_out INTEGER)")
	if err != nil {
		log.Fatal(err)
	}
}

func emailEntryFromRow(row *sql.Row) (*EmailEntry, error) {
	var id int
	var email string
	var confirmedAt int64
	var optOut bool

	err := row.Scan(&id, &email, &confirmedAt, &optOut)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	t := time.Unix(confirmedAt, 0)
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &t, OptOut: optOut}, nil
}

func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec("INSERT INTO emails (email, confirmed_at, opt_out) VALUES (?, 0, false)", email)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	row := db.QueryRow("SELECT id, email, confirmed_at, opt_out FROM emails WHERE email = ?", email)

	emailEntry, err := emailEntryFromRow(row)
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	return emailEntry, nil
}

func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	t := entry.ConfirmedAt.Unix()

	_, err := db.Exec(`INSERT INTO
		emails (id, email, confirmed_at, opt_out)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
		confirmed_at = ?, opt_out = ?`, entry.Id, entry.Email, t, entry.OptOut, t, entry.OptOut)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		UPDATE emails 
		SET opt_out = true 
		WHERE email = ?`, email)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type GetEmailBatchQuery struct {
	Page  int
	Count int
}

func GetEmailBatch(db *sql.DB, params GetEmailBatchQuery) ([]EmailEntry, error) {
	var empty []EmailEntry

	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE opt_out = false
		ORDER BY id ASC
		LIMIT ? OFFSET ?`, params.Count, (params.Page-1)*params.Count)

	if err != nil {
		log.Println(err)
		return empty, err
	}

	defer rows.Close()

	entries := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		var entry EmailEntry
		if err := rows.Scan(&entry.Id, &entry.Email, &entry.ConfirmedAt, &entry.OptOut); err != nil {
			log.Println(err)
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return entries, nil
}
