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
	Address		string
	Email		string
}

var db			*Database
var	services	map[string]*Service
