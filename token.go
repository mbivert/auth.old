package main

import (
	"math/rand"
	"sync"
	"time"
)

const (
	LenToken		=	64
	TokenTimeout	=	3600		// 1h
	alnum 			=	"abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789"
)

type Token struct {
	Key			string		// Key of the service
	Token		string
}

var	utokens		map[int32][]*Token
var	tokens		map[string]int32
var timers		map[string]chan string
var mtoken		*sync.Mutex

func randomString(n int) string {
	buf := make([]byte, n)

	for i := 0; i < LenToken; i++ {
		buf[i] = alnum[rand.Intn(len(alnum))]
	}

	return string(buf)
}

func mkToken() string {
gen:
	token := randomString(LenToken)
	if _, exists := tokens[token]; exists {
		goto gen
	}

	return token
}

func Timeout(token string, update chan string) {
	timeout := time.Tick(TokenTimeout*time.Second)

	for {
		select {
		case <- timeout:
			DelToken(token)
			return
		case token = <- update:
			timeout = time.Tick(TokenTimeout*time.Second)
		}
	}
}

func NewToken(id int32, key string) *Token {
	mtoken.Lock()
		token := mkToken()
		tokens[token] = id
		timers[token] = make(chan string)
		go Timeout(token, timers[token])
		res := &Token{ key, token }
		utokens[id] = append(utokens[id], res)
	mtoken.Unlock()

	return res
}

func CheckToken(token string) bool {
	mtoken.Lock()
		_, ok := tokens[token]
	mtoken.Unlock()

	return ok
}

func DelToken(token string) {
	mtoken.Lock()
		if id := tokens[token]; id > 0 {
			n := len(utokens[id])
			for i := range utokens[id] {
				if utokens[id][i].Token == token {
					utokens[id][i]	= utokens[id][n]
					utokens[id]		= utokens[id][0:n-1]
					break
				}
			}
		}
	mtoken.Unlock()
}

func UpdateToken(token string, key string) string {
	ret := ""

	mtoken.Lock()
		// remove old token
		if id := tokens[token]; id > 0 {
			update := timers[token]
			delete(tokens, token)
	
			// create new one
			ret = mkToken()
			tokens[ret] = id
			timers[ret] = update

			// update timeout
			update <- ret

			// update value
			for i := range utokens[id] {
				if utokens[id][i].Token == token {
					utokens[id][i].Token = ret
					break
				}
			}
		}
	mtoken.Unlock()

	return ret
}
