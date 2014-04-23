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

var this = "http://localhost:8080"

type Token struct {
	Service		string
	Tok			string
}

type Auth struct {
	db			*Database

	connected	map[int32][]Token
	tokens		map[string]int32

//	toact		map[string]bool
}

func NewToken(service string) *Token {
	// maybe use something in nearly whole ascii
	// rather than a-f0-9.
	tok := hex.EncodeToString(
		securecookie.GenerateRandomKey(lenToken))

	return &Token{ service, tok }
}

func NewAuth() *Auth {
	return &Auth{
		NewDB(),
		map[int32][]Token{},
		map[string]int32{},
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

	id, err := a.db.Register(name, email, "citizen")
	if err != nil { return err }

	token := NewToken(this)
	a.StoreToken(id, token)
	err = a.SendToken(email, token)

	return err
}

func (a *Auth) Unregister(name string) {
}

func (a *Auth) Login(name string, token *Token) (ntoken *Token, err error) {
	// user want a new token
	if token.Tok == "" {
		email, err := a.db.GetEmail(name)
		if err != nil { return nil, err }
		id, _ := a.db.GetId(name)

		t := NewToken(this)
		a.StoreToken(id, t)
		err = a.SendToken(email, t)

		return nil, err
	}

//	if a.toact[tok] { a.db.Activate(id); a.toact[tok] = false }

	_, err = a.CheckToken(name, token)
	if err != nil { return nil, err }

	ntoken = NewToken(token.Service)
	a.SetToken(token.Tok, ntoken.Tok)

	return
}

func (a *Auth) Logout(token *Token) {
	id := a.tokens[token.Tok]
	if id == 0 { return }

	delete(a.tokens, token.Tok)

	n := len(a.connected[id])
	for i := range a.connected[id] {
		if a.connected[id][i].Tok == token.Tok {
			a.connected[id][i] = a.connected[id][n-1]
			a.connected[id][n-1] = Token{ "", "" }
			a.connected[id] = a.connected[id][:n-1]
		}
	}
}

func (a *Auth) Update(id int32, name, email string) {
}

func (a *Auth) StoreToken(id int32, token *Token) {
	a.connected[id] = append(a.connected[id], *token)
	a.tokens[token.Tok] = id
}

func (a *Auth) CheckToken(name string, token *Token) (id int32, err error) {
	if len(token.Tok) != lenToken*2 {
		id, err = 0, errors.New("CheckToken: Wrong token length.")
		return
	}

	id = a.tokens[token.Tok]
	id2, _ := a.db.GetId(name)
	if id != id2 || id == 0 { err = errors.New("CheckToken: Wrong token.") }
	return
}

func (a *Auth) QuickCheck(token *Token) (err error) {
	if a.tokens[token.Tok] == 0 {
		err = errors.New("QuickCheck: Invalid Token")
	}
	return
}

func (a *Auth) SetToken(old, new string) {
	id := a.tokens[old]
	delete(a.tokens, old)
	a.tokens[new] = id

	for i := range a.connected[id] {
		if a.connected[id][i].Tok == old {
			a.connected[id][i].Tok = new
		}
	}
}

func (a *Auth) SendToken(email string, token *Token) error {
	err := a.SendEmail(email, "Token", "Hi there,\nHere is your token for "+ token.Service+ ": " + token.Tok)
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

func (a *Auth) GetTokens(token *Token) ([]Token, error) {
	id := a.tokens[token.Tok]
	if id == 0 { return nil, errors.New("GetTokens: Wrong token.") }

	return a.connected[id], nil
}
