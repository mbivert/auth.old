package main

import (
	"math/rand"
)

const (
	LenToken	=	64
	alnum 		=	"abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789"
)

type Token struct {
	Key			string		// Key of the service
	Token		string
}

func RandomString(n int) string {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = alnum[rand.Intn(len(alnum))]
	}
	return string(buf)
}

func NewToken(key string) *Token {
	return &Token{ key, RandomString(LenToken) }
}

func CheckToken(token *Token) bool {
	return tokens[token.Token] != 0
}

func ChainToken(token *Token) (ntoken *Token) {
	if !CheckToken(token) { return nil }

	ntoken = NewToken(token.Key)

	UpdateToken(token.Token, ntoken.Token)

	return
}

func StoreToken(id int32, token *Token) {
	tokens[token.Token] = id
	utokens[id] = append(utokens[id], token)
}

func UpdateToken(old, new string) {
	id := tokens[old]
	delete(tokens, old)
	tokens[new] = id

	for i := range utokens[id] {
		if utokens[id][i].Token == old {
			utokens[id][i].Token = new
		}
	}
}

func DelToken(token *Token) {
	
}
