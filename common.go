package main

type User struct {
	Id			int32
	Name		string
	Email		string
	Admin		bool
}

type Service struct {
	Id			int32
	Name		string
	Url			string
	Key			string
}

var db			*Database
var	services	map[string]*Service

var	utokens		map[int32][]*Token
var	tokens		map[string]int32
