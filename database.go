package main

import (
	_ "github.com/lib/pq"
	"database/sql"
)

var Admin User = User {
		Id		:		1,			// by convention
		Name	:		"admin",
		Email	:		"mathieu.root@gmail.com",
		Admin	:		true,
}
var Auth  Service = Service {
		Id		:		1,			// by convention
		Name	:		"AAS",
		Url		:		"http://auth.awesom.eu",
		Key		:		RandomString(64),
		Address	:		"127.0.0.1",
		Email	:		"mathieu.root@gmail.com",
}


type Database struct {
	*sql.DB
}

// XXX secure connection
func NewDatabase() (db *Database) {
	tmp, err := sql.Open("postgres",
		"dbname=auth user=auth host=localhost sslmode=disable")
	if err != nil {
		LogFatal(err)
	} else {
		db = &Database{ tmp }
		db.Init()
	}

	return
}

func (db *Database) createFatal(descr string) {
	_, err := db.Query(descr)
	if err != nil {
		LogFatal(err)
	}
}

func (db *Database) createTables() {
	db.createFatal(`CREATE TABLE IF NOT EXISTS
		users(
			id						SERIAL,
			name		TEXT		UNIQUE,
			email		TEXT		UNIQUE,
			admin		BOOLEAN,
			PRIMARY KEY ("id")
		)
	`)

	db.createFatal(`CREATE TABLE IF NOT EXISTS
		services(
			id						SERIAL,
			name		TEXT		UNIQUE,
			url			TEXT		UNIQUE,
			key			TEXT		UNIQUE,
			address		INET,
			email		TEXT,
			PRIMARY KEY ("id")
		)
	`)
}

func (db *Database) createAdmin() {
	if db.GetUser(1) == nil {
		if err := db.AddUser(&Admin); err != nil {
			LogFatal(err)
		}
	} else {
		Admin = *db.GetUser(1)
	}
}

func (db *Database) createAuth() {
	if db.GetService(1) == nil {
		if err := db.AddService(&Auth); err != nil {
			LogFatal(err)
		}
	} else {
		Auth = *db.GetService(1)
	}
}

func (db *Database) loadServices() {
	rows, err := db.Query(`
		SELECT id, name, url, key, address, email
		FROM services`)
	if err != nil {
		LogFatal(err)
	}

	for rows.Next() {
		var s Service
		rows.Scan(&s.Id, &s.Name, &s.Url, &s.Key, &s.Address, &s.Email)
		services[s.Key] = &s
	}

}

func (db *Database) Init() {
	db.createTables()
	db.createAdmin()
	db.createAuth()

	db.loadServices()
}

// Users
func (db *Database) AddUser(u *User) error {
	return db.QueryRow(`INSERT INTO
		users(name, email, admin)
		VALUES($1, $2, $3)
		RETURNING id`, u.Name, u.Email, u.Admin).Scan(&u.Id)
}

func (db *Database) GetUser(id int32) *User {
	var u User

	err := db.QueryRow(`
		SELECT id, name, email, admin
		FROM users
		WHERE id = $1`, id).Scan(&u.Id, &u.Name, &u.Email, &u.Admin)

	if err != nil {
		LogError(err)
		return nil
	}

	return &u
}

func (db *Database) GetUser2(login string) *User {
	var u User

	err := db.QueryRow(`
		SELECT id, name, email, admin
		FROM users
		WHERE	name	= $1
		OR		email	= $1`, login).Scan(&u.Id, &u.Name, &u.Email, &u.Admin)

	if err != nil {
		LogError(err)
		return nil
	}

	return &u
}

func (db *Database) GetEmail(name string) (email string) {
	db.QueryRow(`
		SELECT email
		FROM users
		WHERE name = $1`, name).Scan(&email)
	return
}

func (db *Database) IsAdmin(id int32) bool {
	err := db.QueryRow(`
		SELECT id FROM users
		WHERE	id = $1
		AND		admin = true`, id).Scan(&id)

	return err == nil
}

//	DelUser()
//	UpdateUser()
//	Activate()

// Services
func (db *Database) AddService(s *Service) error {
	return db.QueryRow(`INSERT INTO
		services(name, url, key, address, email)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id`, s.Name, s.Url, s.Key, s.Address, s.Email).Scan(&s.Id)
}

func (db *Database) GetService(id int32) *Service {
	var s Service

	err := db.QueryRow(`
		SELECT id, name, url, key, address, email
		FROM services
		WHERE id = $1`, id).Scan(&s.Id, &s.Name, &s.Url,
			&s.Key, &s.Address, &s.Email)

	if err != nil {
		LogError(err)
		return nil
	}

	return &s
}

//	DelService()
//	UpdateService()
