package main

import (
	"errors"
	"flag"
	"github.com/dchest/captcha"
	"html/template"
	"log"
	"net/http"
	"math/rand"
	"time"
	"strconv"
	"strings"
)

var port = flag.String("port", "8080", "Listening HTTP port")

var rtmpl = template.Must(
	template.New("register.html").ParseFiles("templates/register.html"))
var ltmpl = template.Must(
	template.New("login.html").ParseFiles("templates/login.html"))
var stmpl = template.Must(
	template.New("sessions.html").Funcs( template.FuncMap{
		"GetService": func(key string) *Service {
			return services[key]
		},
		}).ParseFiles("templates/sessions.html"))
var atmpl = template.Must(
	template.New("admin.html").Funcs( template.FuncMap{
		"GetUser": func(id int32) *User {
			return db.GetUser(id)
		},
		"GetService": func(key string) *Service {
			return services[key]
		},
		}).ParseFiles("templates/admin.html"))

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		ko(w); return
	}

	token, _ := GetToken(r)	
	writeFiles(w, "templates/header.html", GetNavbar(token),
		"templates/index.html", "templates/footer.html")
}

func getRegister(w http.ResponseWriter, r *http.Request) {
	token, _ := GetToken(r)	
	writeFiles(w, "templates/header.html", GetNavbar(token))

	d := struct { CaptchaId string }{ captcha.New() }

	if err := rtmpl.Execute(w, &d); err != nil {
		LogHttp(w, err); return
	}

	writeFiles(w, "templates/footer.html")
}

func postRegister(w http.ResponseWriter, r *http.Request) {
/*
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaRes")) {
		w.Write([]byte("<p>Bad captcha; try again. </p>"))
		return
	}
*/

	if err := Register(r.FormValue("name"), r.FormValue("email")); err != nil {
		LogHttp(w, err); return
	}

	w.Write([]byte("<p>Check your email account, "+
		`and <a href="/login">login</a>!</p>`))
}

func register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getRegister(w, r)
	case "POST":
		postRegister(w, r)
	default:
		ko(w)
	}
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	writeFiles(w, "templates/header.html", "templates/navbar.html")

	d := struct { CaptchaId string }{ captcha.New() }

	if err := ltmpl.Execute(w, &d); err != nil {
		LogHttp(w, err); return
	}

	writeFiles(w, "templates/footer.html")
}

func postLogin(w http.ResponseWriter, r *http.Request) {
/*
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaRes")) {
		w.Write([]byte("<p>Bad captcha; try again. </p>"))
		return
	}
*/
	token, err := Login(r.FormValue("login"))
	if err != nil {
		LogHttp(w, err)
		return
	}
	if token == "" {
		w.Write([]byte("<p>Check your email account, "+
			`and <a href="/login">login</a>!</p>`))
		return
	}

	err = SetToken(w, token)
	if err != nil { LogHttp(w, err); return }

	http.Redirect(w, r, "/sessions", http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getLogin(w, r)
	case "POST":
		postLogin(w, r)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { ko(w); return }

	token, err := GetToken(r)
	if err != nil { LogHttp(w, err); return }

	Logout(token)
	UnsetToken(w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func admin(w http.ResponseWriter, r *http.Request) {
	token, err := GetToken(r)
	if err != nil || !CheckToken(token) || !IsAdmin(token) {
		ko(w); return
	}

	if r.FormValue("action") == "enable" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		db.SetMode(int32(id), true)
	} else if r.FormValue("action") == "disable" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		db.SetMode(int32(id), false)
	}

	writeFiles(w, "templates/header.html", "templates/navbar3.html")

	// XXX make a copy of utokens/services
	d := struct {
		Users		map[int32][]Token
		Services	map[string]*Service
	}{ utokens, services }

	if err := atmpl.Execute(w, &d); err != nil {
		LogHttp(w, err); return
	}

	writeFiles(w, "templates/footer.html")
}

func getSettings(w http.ResponseWriter, r *http.Request, token string) {
	toks := AllTokens(token)

	writeFiles(w, "templates/header.html", GetNavbar(token))
	d := struct { Tokens []Token }{ toks }
	if err := stmpl.Execute(w, &d); err != nil {
		LogHttp(w, err); return
	}

	writeFiles(w, "templates/footer.html")
}

func postSettings(w http.ResponseWriter, r *http.Request, token string) {
	todel := r.FormValue("token")

	if OwnerToken(todel) == OwnerToken(token) {
		RemoveToken(todel)
	}

	http.Redirect(w, r, "/sessions", http.StatusFound)
}

func sessions(w http.ResponseWriter, r *http.Request) {
	token, err := GetToken(r)
	if err == nil {
		if !CheckToken(token) {
			err = errors.New("Wrong Token")
		}
	}
	if err != nil { LogHttp(w, err); return }

	ntoken := UpdateToken(token)
	err = SetToken(w, ntoken)
	if err != nil { LogHttp(w, err); return }

	switch r.Method {
	case "GET":
		getSettings(w, r, ntoken)
	case "POST":
		postSettings(w, r, ntoken)
	}
}

func discover(w http.ResponseWriter, r *http.Request) {
	name, url := r.FormValue("name"), r.FormValue("url")
	address, email := r.FormValue("address"), r.FormValue("email")

	key, err := AddService(name, url, address, email)
	if err != nil { ko(w); return }

	w.Write([]byte(key))
}

func update(w http.ResponseWriter, r *http.Request) {
	ko(w)
}

func info(w http.ResponseWriter, r *http.Request) {
	token, key := r.FormValue("token"), r.FormValue("key")

	if !CheckService(key, strings.Split(r.RemoteAddr, ":")[0]) {
		ko(w); return
	}

	u := db.GetUser(OwnerToken(token))
	if u == nil { ko(w); return }

	w.Write([]byte(strconv.Itoa(int(u.Id))+"\n"+u.Name+"\n"+u.Email))
}

func alogin(w http.ResponseWriter, r *http.Request) {
	login, key := r.FormValue("login"), r.FormValue("key")

	if !CheckService(key, strings.Split(r.RemoteAddr, ":")[0]) {
		ko(w); return
	}

	if isToken(login) {
		if CheckToken(login) {
			ok(w)
		} else {
			ko(w)
		}
	} else {
		u := db.GetUser2(login)
		if u == nil { ko(w) }
		s := services[key]
		if s == nil { ko(w); return }
		NewToken(u.Id, s.Key)
		w.Write([]byte("new"))
	}
}

func chain(w http.ResponseWriter, r *http.Request) {
	key, token := r.FormValue("key"), r.FormValue("token") 

	if !CheckService(key, strings.Split(r.RemoteAddr, ":")[0]) {
		ko(w); return
	}

	ntoken := UpdateToken(token)
	if ntoken != "" {
		w.Write([]byte(ntoken))
	} else {
		ko(w)
	}
}

func alogout(w http.ResponseWriter, r *http.Request) {
	key, token := r.FormValue("key"), r.FormValue("token") 

	if !CheckService(key, strings.Split(r.RemoteAddr, ":")[0]) {
		ko(w); return
	}

	RemoveToken(token)
	ok(w)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Data init
	utokens = map[int32][]Token{}
	tokens = map[string]int32{}
	timeouts = map[int64][]string{}
	ServiceMode = Manual

	go ProcessMsg()
	go Timeouts()

	db = NewDatabase()

	// Auth website
	http.HandleFunc("/", index)
	http.HandleFunc("/register/", register)
//	http.HandleFunc("/unregister/", unregister) // TODO
	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", logout)
	http.HandleFunc("/admin/", admin)
	http.HandleFunc("/sessions/", sessions)

	// API
	http.HandleFunc("/api/discover", discover)
	http.HandleFunc("/api/update", update)
	http.HandleFunc("/api/info", info)
	http.HandleFunc("/api/login", alogin)
	http.HandleFunc("/api/chain", chain)
	http.HandleFunc("/api/logout", alogout)

	// Captchas
	http.Handle("/captcha/",
		captcha.Server(captcha.StdWidth, captcha.StdHeight))

	// Static files
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static/"))))

	log.Println("Launching on http://localhost:"+*port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
