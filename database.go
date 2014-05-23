package main

import (
	_ "github.com/lib/pq"
	"database/sql"
)

// Remaining fields loaded from config
var Admin User = User {
		Id		:		1,			// by convention
		Name	:		"admin",
		Admin	:		true,
}
var Auth Service = Service {
		Id		:		1,			// by convention
		Mode	:		true,		// activated
		Address	:		"127.0.0.1",
}

type Database struct {
				*sql.DB
	services	map[string]*Service
}

func NewDatabase() (*Database, error) {
	tmp, err := sql.Open("postgres", C.DBConnect)
	if err != nil { return nil, Err(err) }

	db = &Database{ tmp, map[string]*Service{} }

	return db, db.Init()
}

func (db *Database) createTables() error {
	if _, err := db.Query(`CREATE TABLE IF NOT EXISTS
		users(
			id						SERIAL,
			name		TEXT		UNIQUE,
			email		TEXT		UNIQUE,
			passwd		TEXT,
			admin		BOOLEAN,
			PRIMARY KEY ("id")
		)
	`); err != nil { return Err(err) }

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
	`); err != nil { return Err(err) }

	return nil
}

func (db *Database) createAdmin() error {
	Admin.Email = C.AdminEmail

	if u, err := db.GetUser(1); err != nil {
		return db.AddUser(&Admin)
	} else {
		Admin = *u
	}

	return nil
}

func (db *Database) createAuth() error {
	Auth.Name = C.Name
	Auth.Url = C.URL
	Auth.Email =  C.AdminEmail
	Auth.Key = randomString(C.LenKey)

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
		return Err(err)
	}

	for rows.Next() {
		var s Service
		rows.Scan(&s.Id, &s.Name, &s.Url, &s.Key, &s.Mode, &s.Address, &s.Email)
		db.services[s.Key] = &s
	}

	return nil
}

func (db *Database) Init() error {
	if err := db.createTables(); err != nil { return err }
	if err := db.createAdmin(); err != nil { return err }
	if err := db.createAuth(); err != nil { return err }

	return db.loadServices()
}

// Users
func (db *Database) AddUser(u *User) error {
	if err := db.QueryRow(`INSERT INTO
		users(name, email, passwd, admin)
		VALUES($1, $2, $3, $4)
		RETURNING id`, u.Name, u.Email, u.Passwd,
			u.Admin).Scan(&u.Id); err != nil {
		return Err(err)
	}

	return nil
}

func (db *Database) GetUser(id int32) (*User, error) {
	var u User

	if err := db.QueryRow(`
		SELECT id, name, email, passwd, admin
		FROM users
		WHERE id = $1`, id).Scan(&u.Id, &u.Name,
			&u.Email, &u.Passwd, &u.Admin); err != nil {
		return nil, Err(err)
	}

	return &u, nil
}

func (db *Database) GetUser2(login string) (*User, error) {
	var u User

	if err := db.QueryRow(`
		SELECT id, name, email, passwd, admin
		FROM users
		WHERE	name	= $1
		OR		email	= $1`, login).Scan(&u.Id,
			&u.Name, &u.Email, &u.Passwd, &u.Admin); err != nil {
		return nil, Err(err)
	}

	return &u, nil
}

func (db *Database) GetUsers() (users []User) {
	rows, _ := db.Query(`SELECT id, name, email, admin
		FROM users`)

	for rows.Next() {
		var u User
		rows.Scan(&u.Id, &u.Name, &u.Email, &u.Admin)
		users = append(users, u)
	}	

	return
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

// Toggle admin status for any user but Admin
func (db *Database) ToggleAdmin(id int32) {
	if id != Admin.Id {
		db.Query(`UPDATE users
			SET admin = NOT admin
			WHERE id = $1`, id)
	}
}

// return admins emails
func (db *Database) GetAdmins() ([]string, error) {
	var emails []string

	rows, err := db.Query(`SELECT email
		FROM users
	WHERE admin = true`)

	if err != nil { return nil, Err(err) }

	for rows.Next() {
		var email string
		rows.Scan(&email)
		emails = append(emails, email)
	}

	return emails, nil
}

func (db *Database) DelUser(id int32) (email string) {
	db.QueryRow(`DELETE FROM users WHERE id = $1
		RETURNING email`, id).Scan(&email)
	return
}

// update name and email
func (db *Database) UpdateName(id int32, name string) error {
	_, err := db.Query(`UPDATE users
		SET name = $2 WHERE id = $1`, id, name)

	return err
}

func (db *Database) UpdateEmail(id int32, email string) error {
	_, err := db.Query(`UPDATE users
		SET email = $2 WHERE id = $1`, id, email)

	return err
}

func (db *Database) UpdatePassword(id int32, passwd string) {
	db.Query(`UPDATE users SET passwd = $2 WHERE id = $1`, id, passwd)
}

//	Activate()

// Services
func (db *Database) AddService(s *Service) error {
	err := db.QueryRow(`INSERT INTO
		services(name, url, key, mode, address, email)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id`, s.Name, s.Url, s.Key, s.Mode, s.Address, s.Email).Scan(&s.Id)

	if err != nil {	return Err(err) }

	db.services[s.Key] = s

	return nil
}

func (db *Database) GetService(id int32) (*Service, error) {
	var s Service

	err := db.QueryRow(`
		SELECT id, name, url, key, mode, address, email
		FROM services
		WHERE id = $1`, id).Scan(&s.Id, &s.Name, &s.Url,
			&s.Key, &s.Mode, &s.Address, &s.Email)

	if err != nil { return nil, Err(err) }

	return &s, nil
}

func (db *Database) GetService2(key string) *Service {
	return db.services[key]
}

func (db *Database) GetServices() map[string]*Service {
	return db.services
}

func (db *Database) SetMode(id int32, on bool) error {
	var key string

	if id == Auth.Id { return Err(NonSense) }

	err := db.QueryRow(`
		UPDATE services
			SET mode = $1
			WHERE id = $2
		RETURNING key`, on, id).Scan(&key)
	if err != nil { return Err(err) }

	db.services[key].Mode = on

	return nil
}

func (db *Database) DelService(key string) {
	if key != Auth.Key {
		db.Query("DELETE FROM services WHERE key = $1", key)
		delete(db.services, key)
		// XXX delete associated tokens
	}
}

func (db* Database) UpdateKey(key, nkey string) {
	// if collision happens, it MUST be a sign or something, so let it be
	db.services[nkey] = db.services[key]
	delete(db.services, key)
	db.Query("UPDATE services SET key = $1 WHERE key = $2", nkey, key)
}

