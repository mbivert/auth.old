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
	Mode		bool
	Address		string
	Email		string
}

var db			*Database
var	services	map[string]*Service

const (
	Automatic = iota
	Manual
	Disabled
)

var ServiceMode int
