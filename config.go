package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	URL				string
	Port			string
	Name			string
	AdminEmail		string
	DBConnect		string

	Mode			string
	Timeout			int64
	LenToken		int
	LenKey			int

	VerifyCaptcha	bool

	SSL				bool
	Certificate		string
	PKey			string

	SMTPServer		string
	SMTPPort		string
	AuthEmail		string
	AuthPasswd		string
}

var C Config

func LoadConfig(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Cannot read configuration file: ", err)
	}

	if err := json.Unmarshal(data, &C); err != nil {
		log.Fatal("Error while parsing configuration file: ", err)
	}

	switch C.Mode {
	case "Disable": ServiceMode = Disabled
	case "Automatic": ServiceMode = Automatic
	default: ServiceMode = Manual
	}

	// No checking:
	// Wrong configuration -> undefined behavior.
}
