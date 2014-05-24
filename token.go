package main

import (
	"time"
	//	"log"
	//	"reflect"
)

type Token struct {
	Key   string // Key of the service
	Token string
}

var utokens = map[int32][]Token{}
var tokens = map[string]int32{}
var timeouts = map[int64][]string{}

var chanmsg chan Msg

type Msg interface {
	process()
}

func mkToken() string {
gen:
	token := randomString(C.LenToken)
	if _, exists := tokens[token]; exists {
		goto gen
	}

	return token
}

type NewMsg struct {
	uid    int32
	key    string
	answer chan string
}

func (m NewMsg) process() {
	token := mkToken()
	// store token
	tokens[token] = m.uid
	utokens[m.uid] = append(utokens[m.uid], Token{m.key, token})

	// setup timeout
	exptime := time.Now().Unix() + C.Timeout
	timeouts[exptime] = append(timeouts[exptime], token)

	// return value
	m.answer <- token
}

type RemoveMsg struct {
	token string
}

func (m RemoveMsg) process() {
	if id := tokens[m.token]; id > 0 {
		n := len(utokens[id])
		for i := range utokens[id] {
			if utokens[id][i].Token == m.token {
				utokens[id][i] = utokens[id][n-1]
				utokens[id] = utokens[id][0 : n-1]
				break
			}
		}
		delete(tokens, m.token)
	}
}

type CheckMsg struct {
	token  string
	answer chan bool
}

func (m CheckMsg) process() {
	_, ok := tokens[m.token]
	m.answer <- ok
}

type UpdateMsg struct {
	token  string
	answer chan string
}

func (m UpdateMsg) process() {
	// check old one
	id, ok := tokens[m.token]
	if !ok {
		m.answer <- ""
		return
	}

	token := mkToken()

	delete(tokens, m.token)

	// create new one
	tokens[token] = id
	// update value
	for i := range utokens[id] {
		if utokens[id][i].Token == m.token {
			utokens[id][i].Token = token
			break
		}
	}

	// setup timeout
	exptime := time.Now().Unix() + C.Timeout
	timeouts[exptime] = append(timeouts[exptime], token)

	// return new one
	m.answer <- token
}

type AllMsg struct {
	token  string
	answer chan []Token
}

func (m AllMsg) process() {
	m.answer <- utokens[tokens[m.token]]
}

type OwnMsg struct {
	token  string
	answer chan int32
}

func (m OwnMsg) process() {
	m.answer <- tokens[m.token]
}

// background processes
func ProcessMsg() {
	chanmsg = make(chan Msg)
	for {
		m := <-chanmsg
		//		log.Println("Process: ", reflect.TypeOf(m), ", ", m)
		m.process()
	}
}
func Timeouts() {
	for {
		time.Sleep(2 * time.Second)
		now := time.Now().Unix()
		for date, toks := range timeouts {
			if date <= now {
				for _, token := range toks {
					RemoveToken(token)
				}
				delete(timeouts, date)
			}
		}
	}
}

// "API"
func NewToken(uid int32, key string) *Token {
	answer := make(chan string, 1)
	chanmsg <- NewMsg{uid, key, answer}

	return &Token{key, <-answer}
}

func CheckToken(token string) bool {
	answer := make(chan bool, 1)
	chanmsg <- CheckMsg{token, answer}

	return <-answer
}

func UpdateToken(token string) string {
	answer := make(chan string, 1)
	chanmsg <- UpdateMsg{token, answer}

	return <-answer
}

func RemoveToken(token string) {
	chanmsg <- RemoveMsg{token}
}

func AllTokens(token string) []Token {
	answer := make(chan []Token, 1)

	chanmsg <- AllMsg{token, answer}

	return <-answer
}

func OwnerToken(token string) int32 {
	answer := make(chan int32, 1)

	chanmsg <- OwnMsg{token, answer}

	return <-answer
}
