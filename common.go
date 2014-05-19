package main

import (
	"errors"
	"runtime"
	"strconv"
	"time"
)

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

var ServiceMode = Manual

type InternalError struct {
	Date		time.Time
	File		string
	Line		int
	Msg			string
}

func (e *InternalError) Error() string {
	return e.Date.String()+" "+e.File+":"+strconv.Itoa(e.Line)+
		" "+e.Msg
}

func MkIErr(err error) *InternalError {
	_, file, line, _ := runtime.Caller(1)

	return &InternalError{ time.Now(), file, line, err.Error() }
}

var (
	NoSuchErr		= errors.New("No such name/email")
	NoSuchTErr		= errors.New("No such token")
	NonSense		= errors.New("This Sense Makes No Action")
	NoEmailErr		= errors.New("Email is mandatory")
	LongEmailErr	= errors.New("Email is too long (maxsize: "+
						strconv.Itoa(C.LenToken-1)+")")
	EmailFmtErr		= errors.New("Wrong Email format (you@provider)")
	NoNameErr		= errors.New("Name is mandatory")
	LongNameErr		= errors.New("Name is too long (maxsize: "+
						strconv.Itoa(C.LenToken-1)+")")
	NameFmtErr		= errors.New("Invalid characters in name (no whites or @)")
	WrongUser		= errors.New("User name or password already in use")
	SMTPErr			= errors.New("Email not send. Contact an admin.")
	MouldyCookie	= errors.New("Mouldy Cookie, Sour Tea!")
	NotAdminErr		= errors.New("Can't go there.")
	SetCookieErr	= errors.New("Can't set cookie. (contact us)")
	EmptyFieldsErr	= errors.New("void.")
	BadCaptchaErr	= errors.New("Bad Captcha. Try again")
)
