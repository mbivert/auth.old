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
		Key		:		randomString(64),
		Mode	:		true,
		Address	:		"127.0.0.1",
		Email	:		"mathieu.root@gmail.com",
}


type Database struct {
	*sql.DB
}

// XXX secure connection
func NewDatabase() (*Database, error) {
	tmp, err := sql.Open("postgres",
		"dbname=auth user=auth host=localhost sslmode=disable")
	if err != nil { return nil, MkIErr(err) }

	db = &Database{ tmp }

	return db, db.Init()
}

func (db *Database) createTables() error {
	if _, err := db.Query(`CREATE TABLE IF NOT EXISTS
		users(
			id						SERIAL,
			name		TEXT		UNIQUE,
			email		TEXT		UNIQUE,
			admin		BOOLEAN,
			PRIMARY KEY ("id")
		)
	`); err != nil { return MkIErr(err) }

	if _, err := db.Query(`CREATE TABLE IF NOT EXISTS
		services(
			id						SERIAL,
			name		TEXT		UNIQUE,
			url			TEXT		UNIQUE,
			key			TEXT		UNIQUE,
			mode		BOOLEAN,
			address		INET,
			email		TEXT,
			PRIMARY KEY ("id")
		)
	`); err != nil { return MkIErr(err) }

	return nil
}

func (db *Database) createAdmin() error {
	if u, err := db.GetUser(1); err != nil {
		return db.AddUser(&Admin)
	} else {
		Admin = *u
	}

	return nil
}

func (db *Database) createAuth() error {
	if s, err := db.GetService(1); err != nil {
		return db.AddService(&Auth)
	} else {
		Auth = *s
	}

	return nil
}

func (db *Database) loadServices() error {
	rows, err := db.Query(`
		SELECT id, name, url, key, mode, address, email
		FROM services`)
	if err != nil {
		return MkIErr(err)
	}

	for rows.Next() {
		var s Service
		rows.Scan(&s.Id, &s.Name, &s.Url, &s.Key, &s.Mode, &s.Address, &s.Email)
		services[s.Key] = &s
	}

	return nil
}

func (db *Database) Init() error {
	services = map[string]*Service{}

	if err := db.createTables(); err != nil { return err }
	if err := db.createAdmin(); err != nil { return err }
	if err := db.createAuth(); err != nil { return err }

	return db.loadServices()
}

// Users
func (db *Database) AddUser(u *User) error {
	if err := db.QueryRow(`INSERT INTO
		users(name, email, admin)
		VALUES($1, $2, $3)
		RETURNING id`, u.Name,
			u.Email, u.Admin).Scan(&u.Id); err != nil {
		return MkIErr(err)
	}

	return nil
}

func (db *Database) GetUser(id int32) (*User, error) {
	var u User

	if err := db.QueryRow(`
		SELECT id, name, email, admin
		FROM users
		WHERE id = $1`, id).Scan(&u.Id, &u.Name,
			&u.Email, &u.Admin); err != nil {
		return nil, MkIErr(err)
	}

	return &u, nil
}

func (db *Database) GetUser2(login string) (*User, error) {
	var u User

	if err := db.QueryRow(`
		SELECT id, name, email, admin
		FROM users
		WHERE	name	= $1
		OR		email	= $1`, login).Scan(&u.Id,
			&u.Name, &u.Email, &u.Admin); err != nil {
		return nil, MkIErr(err)
	}

	return &u, nil
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

func (db *Database) GetAdminMail() ([]string, error) {
	var emails []string

	rows, err := db.Query(`SELECT email
		FROM users
	WHERE admin = true`)

	if err != nil { return nil, MkIErr(err) }

	for rows.Next() {
		var email string
		rows.Scan(&email)
		emails = append(emails, email)
	}

	return emails, nil
}

//	DelUser()
//	UpdateUser()
//	Activate()

// Services
func (db *Database) AddService(s *Service) error {
	err := db.QueryRow(`INSERT INTO
		services(name, url, key, mode, address, email)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id`, s.Name, s.Url, s.Key, s.Mode, s.Address, s.Email).Scan(&s.Id)

	if err != nil {	return MkIErr(err) }

	services[s.Key] = s

	return nil
}

func (db *Database) GetService(id int32) (*Service, error) {
	var s Service

	err := db.QueryRow(`
		SELECT id, name, url, key, mode, address, email
		FROM services
		WHERE id = $1`, id).Scan(&s.Id, &s.Name, &s.Url,
			&s.Key, &s.Mode, &s.Address, &s.Email)

	if err != nil { return nil, MkIErr(err) }

	return &s, nil
}

func (db *Database) GetService2(key string) *Service {
	return services[key]
}

func (db *Database) SetMode(id int32, on bool) error {
	var key string

	if id == Auth.Id { return MkIErr(NonSense) }

	err := db.QueryRow(`
		UPDATE services
			SET mode = $1
			WHERE id = $2
		RETURNING key`, on, id).Scan(&key)
	if err != nil { return MkIErr(err) }

	services[key].Mode = on

	return nil
}

//	DelService()
//	UpdateService()
