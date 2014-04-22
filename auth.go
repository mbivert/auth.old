package main

// XXX clean this up. make cache for email<->id

import (
	"errors"
	"github.com/gorilla/securecookie"
	"encoding/hex"
	"net/smtp"
	"strings"
)

const (
	lenToken	=	32
)

type Auth struct {
	db *Database

	connected map[int32][]string
	tokens map[string]int32

//	toact map[string]bool
}

func NewAuth() *Auth {
	return &Auth{
		NewDB(),
		make(map[int32][]string),
		make(map[string]int32),
//		make(map[string]bool),
	}
}

func (a *Auth) Register(name, email string) error {
	if name == "" || email == "" {
		err := errors.New("Empty field(s)")
		return err
	}

	// don't even bother.
	if !strings.Contains(email, "@") {
		err := errors.New("Bad email address")
		return err
	}

	_, err := a.db.Register(name, email, "citizen")
	if err != nil {
		return err
	}

	err = a.SendToken(email)
	return err
}

func (a *Auth) Unregister(name string) {
}

func (a *Auth) Login(name, tok string) (token string, err error) {
	// user want a new token
	if tok == "" {
		email, err := a.db.GetEmail(name)
		if err == nil { a.SendToken(email) }
		return "", err
	}
	// *2 because it has been hex.encoded()
	if len(tok) != lenToken*2 || a.tokens[tok] == 0 {
		err = errors.New("Wrong token.")
		return "", err
	}

	id := a.tokens[tok]
	id2, _ := a.db.GetId(name)

	if id != id2 {
		err = errors.New("Wrong token.")
		return "", err
	}

//	if a.toact[tok] { a.db.Activate(id); a.toact[tok] = false }

	// Create a fresh token
	delete(a.tokens, tok)
	email, _ := a.db.GetEmail(name)
	token, err = a.MkToken(email)

	return
}

func (a *Auth) Logout(tok string) {
}

func (a *Auth) Update(id int32, name, email string) {
}

func (a *Auth) MkToken(email string) (string, error) {
	id, err := a.db.GetId(email)
	if err != nil { return "", err }

	// maybe use something in nearly whole ascii
	// rather than a-f0-9.
	tok := hex.EncodeToString(
		securecookie.GenerateRandomKey(lenToken))

	a.connected[id] = append(a.connected[id], tok)
	a.tokens[tok] = id

	return tok, nil
}

func (a *Auth) SendToken(email string) error {
	tok, err := a.MkToken(email)
	if err != nil { return err }

	err = a.SendEmail(email, "Token", "Here is your token: "+tok)
	return err
}

// XXX Change SMTP server to smtp.awesom.eu
func (a *Auth) SendEmail(to, subject, msg string) error {
	from := "auth.newsome@gmail.com"
	passwd := "awesom auth server"

	body := "To: " + to + "\r\nSubject: " +
		subject + "\r\n\r\n" + msg

	auth := smtp.PlainAuth("", from, passwd, "smtp.gmail.com")

	err := smtp.SendMail("smtp.gmail.com:587", auth, from,
		[]string{to},[]byte(body))

	return err
}
