package main

import (
	_ "github.com/lib/pq"
	"database/sql"
	"errors"
	"log"
)

type Database struct {
	*sql.DB
}

// XXX add password; check for clean user setup; check SSL
func NewDB() (db *Database) {
	tmp, err := sql.Open("postgres",
		"dbname=auth user=auth host=localhost sslmode=disable")
	if err != nil {
		log.Fatal("PostgreSQL connection error:", err)
	} else {
		db = &Database{ tmp }
		db.CreateTables()
	}

	return
}

func (db *Database) CreateTables() {
	row, err := db.Query("SELECT 1 FROM pg_type WHERE typname = 'utype'")
	// if utype doesn't exist, create it.
	if err == nil && !row.Next() {
		_, err = db.Query(`CREATE TYPE utype AS ENUM
			(
				'admin',
				'citizen'
			)
		`)
	}
	if err != nil {
		log.Fatal("Creation of utype failed:", err)
	}

	_, err = db.Query(`CREATE TABLE IF NOT EXISTS
		users(
			id						SERIAL,
			name		TEXT		UNIQUE,
			email		TEXT		UNIQUE,
			type		UTYPE,
			activated	BOOLEAN,
			PRIMARY KEY ("id")
		)
	`)
	if err != nil {
		log.Fatal("Creation of table users failed:", err)
	}
}

type User struct {
	Id			int32
	Name		string
	Email		string
	Type		string
}

func (db *Database) GetEmail(login string) (email string, err error) {
	err = db.QueryRow(`
		SELECT email FROM
			users
		WHERE
				name	= $1
			OR	email	= $1`, login).Scan(&email)
	if err != nil { err = errors.New("Email not found.") }
	return
}

func (db *Database) GetId(login string) (id int32, err error) {
	err = db.QueryRow(`
		SELECT id FROM
			users
		WHERE
				name	= $1
			OR	email	= $1`, login).Scan(&id)
	if err != nil { err = errors.New("Wrong login: "+login) }
	return
}

func (db *Database) Activate(id int32) {
	db.Query(`
		UPDATE users
		SET		activated = true
		WHERE	id = $1`, id)
}

func (db *Database) Register(name, email, typ string) (int32, error) {
	var id int32
	err := db.QueryRow(`INSERT INTO
		users(name, email, type, activated)
		VALUES($1, $2, $3, false)
		RETURNING id`, name, email, typ).Scan(&id)
	if err != nil {
		// check if either email or name was wrong
		var tmp string
		err = db.QueryRow(`SELECT id FROM users
			WHERE name = $1`, name).Scan(&tmp)
		if err == nil {
			return 0, errors.New("Sorry, name already used")
		} else {
			return 0, errors.New("Sorry, email already used")
		}
	}

	return id, nil
}

/*
func (db *Database) UpdateUser(u *User) {
	_, err := db.Query(`UPDATE users
			SET(name, type) = ($1, $2)
		WHERE
				id		= $3
			AND email	= $4`, u.Name, u.Type, u.Id, , u.Email)
	if err != nil {
		log.Println("Error while updating user with id", u.Id, ":", err)
	}
}

func (db *Database) Unregister(id int32) {
	_, err := db.Query("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Println("Error while deleting user with id", id, ":", err)
	}
}

func (db *Database) IsType(id int32, t string) bool {
	 err := db.QueryRow(`SELECT id FROM users
	 	WHERE id = $1 AND type = $2`, id, t).Scan(&id)

	 return err == nil
}
*/
