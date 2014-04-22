package main

import (
	"flag"
	"github.com/dchest/captcha"
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"log"
	"net/http"
	"html/template"
)

var port = flag.String("port", "8080", "Listening HTTP port")

var rtmpl = template.Must(
	template.New("register.html").ParseFiles("templates/register.html"))
var ltmpl = template.Must(
	template.New("login.html").ParseFiles("templates/login.html"))

var A *Auth

var hashKey = []byte(securecookie.GenerateRandomKey(32))
var blockKey = []byte(securecookie.GenerateRandomKey(32))
var s = securecookie.New(hashKey, blockKey)

// Utilities

func writeFiles(w http.ResponseWriter, fs ...string) error {
	for _, f := range fs {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		w.Write(b)
	}
	return nil
}

func no(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("no."))
}

func setToken(w http.ResponseWriter, token string) error {
	encoded, err := s.Encode("auth-token", token) 
	if err == nil {
		cookie := &http.Cookie{
			Name:	"auth-token",
			Value:	encoded,
			Path:	"/",
		}
		http.SetCookie(w, cookie)
	}

	return err
}

func getToken(r *http.Request) (token string, err error) {
	cookie, err := r.Cookie("auth-token")
	if err != nil { return "", err }

	err = s .Decode("auth-token", cookie.Value, &token)
	if err != nil { return "", err }

	return token, nil
}

// Registration

func getRegister(w http.ResponseWriter, r *http.Request) {
	writeFiles(w, "templates/header.html", "templates/navbar.html")

	d := struct { CaptchaId string }{ captcha.New() }

	if err := rtmpl.Execute(w, &d); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	name, email := r.FormValue("name"), r.FormValue("email")

	err := A.Register(name, email)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`<p>Check your email account, and <a href="/login">login</a>!</p>`))
}

func register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getRegister(w, r)
	case "POST":
		postRegister(w, r)
	default:
		no(w)
	}
}

// Login

func getLogin(w http.ResponseWriter, r *http.Request) {
	writeFiles(w, "templates/header.html", "templates/navbar.html")

	d := struct { CaptchaId string }{ captcha.New() }

	if err := ltmpl.Execute(w, &d); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	name, token := r.FormValue("name"), r.FormValue("token")

	tok, err := A.Login(name, token)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = setToken(w, tok)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getLogin(w, r)
	case "POST":
		postLogin(w, r)
	default:
		no(w)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		no(w)
	}

	token, err := getToken(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	A.Logout(token)
}

// Settings

func getSettings(w http.ResponseWriter, r *http.Request) {
	writeFiles(w, "templates/header.html", "templates/navbar2.html",
		"templates/settings.html", "templates/footer.html")
}

func postSettings(w http.ResponseWriter, r *http.Request) {
}

func settings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getSettings(w, r)
	case "POST":
		postSettings(w, r)
	default:
		no(w)
	}
}

// API
func getinfo(w http.ResponseWriter, r *http.Request) {
}

func chain(w http.ResponseWriter, r *http.Request) {
}

// Index

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		no(w)
	}

	writeFiles(w, "templates/header.html", "templates/navbar.html",
		"templates/index.html", "templates/footer.html")
}

func main() {
	flag.Parse()

	A = NewAuth()

	http.HandleFunc("/", index)
	http.HandleFunc("/register/", register)
	http.HandleFunc("/login/", login)

	http.HandleFunc("/logout/", logout)

	http.HandleFunc("/api/getinfo", getinfo)
	http.HandleFunc("/api/chain", chain)

	http.HandleFunc("/settings/", settings)

	http.Handle("/captcha/",
		captcha.Server(captcha.StdWidth, captcha.StdHeight))

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static/"))))


	log.Println("Launching on http://localhost:"+*port)
//	authserver = "http://localhost:"+*port
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
