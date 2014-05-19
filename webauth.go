package main

import (
	"flag"
	"github.com/dchest/captcha"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	conf = flag.String("conf", "config.json", "Configuration file")

	loginForm []byte

	rtmpl = template.Must(
		template.New("register.html").ParseFiles("templates/register.html"))

	stmpl = template.Must(
		template.New("sessions.html").Funcs(template.FuncMap{
			"GetService": func(key string) *Service {
				return db.GetService2(key)
			},
		}).ParseFiles("templates/sessions.html"))

	atmpl = template.Must(
		template.New("admin.html").ParseFiles("templates/admin.html"))

	ntmpl = template.Must(
		template.New("navbar.html").ParseFiles("templates/navbar.html"))
)

func index(w http.ResponseWriter, r *http.Request, token string) {
	writeFiles(w, "templates/index.html")
}

func register(w http.ResponseWriter, r *http.Request, token string) {
	switch r.Method {
	case "GET":
		d := struct{ CaptchaId string }{captcha.New()}
		if err := rtmpl.Execute(w, &d); err != nil {
			log.Println(err)
		}
	case "POST":
		if C.VerifyCaptcha {
			if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaRes")) {
				SetError(w, BadCaptchaErr)
				http.Redirect(w, r, "/register", http.StatusFound)
				return
		}
		}

		if err := Register(r.FormValue("name"), r.FormValue("email")); err != nil {
			SetError(w, err)
			http.Redirect(w, r, "/register", http.StatusFound)
			return
		}

		SetInfo(w, "Check your email account!")
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func login(w http.ResponseWriter, r *http.Request, token string) {
	switch r.Method {
	case "GET":
		w.Write(loginForm)
	case "POST":
		if token, err := Login(r.FormValue("login")); err != nil {
			SetError(w, err)
		} else if token == "" {
			SetInfo(w, "Check you email account!")
		} else if err := SetToken(w, token); err != nil {
			log.Println(err)
			SetError(w, SetCookieErr)
		} else {
			http.Redirect(w, r, "/sessions", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func logout(w http.ResponseWriter, r *http.Request, token string) {
	Logout(token)
	UnsetToken(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func admin(w http.ResponseWriter, r *http.Request, token string) {
	if r.Method == "POST" {
		switch r.FormValue("action") {
		case "enable":
			id, _ := strconv.Atoi(r.FormValue("id"))
			db.SetMode(int32(id), true)
		case "disable":
			id, _ := strconv.Atoi(r.FormValue("id"))
			db.SetMode(int32(id), false)
		case "add":
			_, err := AddService(r.FormValue("name"), r.FormValue("url"),
				r.FormValue("address"), r.FormValue("email"))
			// XXX makes sense to enable service here too
			if err != nil {
				SetError(w, err)
			}
		case "mode-auto":
			ServiceMode = Automatic
			SendAdmin("[AAS] Automatic mode enabled", "Hope you're debugging.")
		case "mode-manual":
			ServiceMode = Manual
		case "mode-disabled":
			ServiceMode = Disabled
		}
		http.Redirect(w, r, "/admin", http.StatusFound)
	}

	d := struct {
		Services map[string]*Service
	}{services}

	if err := atmpl.Execute(w, &d); err != nil {
		log.Println(err)
	}
}

func sessions(w http.ResponseWriter, r *http.Request, token string) {
	switch r.Method {
	case "GET":
		d := struct{ Tokens []Token }{AllTokens(token)}
		if err := stmpl.Execute(w, &d); err != nil {
			log.Println(err)
		}
	case "POST":
		todel := r.FormValue("token")
		if OwnerToken(todel) == OwnerToken(token) {
			RemoveToken(todel)
		}

		http.Redirect(w, r, "/sessions", http.StatusFound)
	}
}

var authfuncs = map[string]func(http.ResponseWriter, *http.Request, string){
	"":         index,
	"register": register,
	//	"unregister":		unregister,
	"login":    login,
	"logout":   logout,
	"admin":    admin,
	"sessions": sessions,
}

// pages which requires to be authenticated
var mustauth = map[string]bool {
	"unregister":true,
	"logout":true,
	"admin":true,
	"sessions":true,
}

func auth(w http.ResponseWriter, r *http.Request) {
	var token string

	f := r.URL.Path[1:] // skip '/'
	// remove trailing '/' if any
	if len(f) != 0 && f[len(f)-1] == '/' {
		f = f[:len(f)-1]
	}

	if authfuncs[f] == nil {
		http.NotFound(w, r)
		return
	}

	// Verify token is valid
	var err error
	if token, err = VerifyToken(r); err != nil && mustauth[f] {
		SetError(w, err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Check permission
	if f == "admin" && !IsAdmin(token) {
		SetError(w, NotAdminErr)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Generate a new token if needed
	if token != "" {
		token = UpdateToken(token)
		if err := SetToken(w, token); err != nil {
			log.Println(err)
			SetError(w, SetCookieErr)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	if r.Method == "GET" && f != "logout" {
		writeFiles(w, "templates/header.html")
		d := struct{ Connected, Admin bool }{
			Connected: token != "",
			Admin:     IsAdmin(token),
		}
		if err := ntmpl.Execute(w, &d); err != nil {
			log.Println(err)
		}
	}

	authfuncs[f](w, r, token)

	if r.Method == "GET" && f != "logout" {
		writeFiles(w, "templates/footer.html")
	}
}

func discover(w http.ResponseWriter, r *http.Request) {
	name, url := r.FormValue("name"), r.FormValue("url")
	address, email := r.FormValue("address"), r.FormValue("email")

	key, err := AddService(name, url, address, email)
	if err != nil {
		ko(w)
		return
	}

	w.Write([]byte(key))
}

func update(w http.ResponseWriter, r *http.Request) {
	ko(w)
}

func info(w http.ResponseWriter, r *http.Request) {
	token, key := r.FormValue("token"), r.FormValue("key")

	if !CheckService(key, strings.Split(r.RemoteAddr, ":")[0]) {
		ko(w)
		return
	}

	if u, err := db.GetUser(OwnerToken(token)); err != nil {
		ko(w)
	} else {
		w.Write([]byte(strconv.Itoa(int(u.Id)) + "\n" + u.Name + "\n" + u.Email))
	}
}

func login2(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")

	if isToken(login) {
		if CheckToken(login) {
			ok(w)
		} else {
			ko(w)
		}
	} else {
		if u, err := db.GetUser2(login); err != nil {
			ko(w)
		} else {
			NewToken(u.Id, db.GetService2(r.FormValue("key")).Key)
			w.Write([]byte("new"))
		}
	}
}

func chain(w http.ResponseWriter, r *http.Request) {
	ntoken := UpdateToken(r.FormValue("token"))
	if ntoken != "" {
		w.Write([]byte(ntoken))
	} else {
		ko(w)
	}
}

func logout2(w http.ResponseWriter, r *http.Request) {
	RemoveToken(r.FormValue("token"))
	ok(w)
}

var apifuncs = map[string]func(http.ResponseWriter, *http.Request){
	"discover": discover,
	"update":   update,
	"info":     info,
	"login":    login2,
	"chain":    chain,
	"logout":   logout2,
}

func api(w http.ResponseWriter, r *http.Request) {
	f := r.URL.Path[5:] // skip '/api/'
	if f != "discover" {
		key := r.FormValue("key")
		if !CheckService(key, strings.Split(r.RemoteAddr, ":")[0]) {
			ko(w)
			return
		}
	}
	if apifuncs[f] != nil {
		apifuncs[f](w, r)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	// XXX do it before anything else.
	LoadConfig(*conf)

	go ProcessMsg()
	go Timeouts()

	var err error
	if db, err = NewDatabase(); err != nil {
		log.Fatal(err)
	}

	if loginForm, err = ioutil.ReadFile("templates/login.html"); err != nil {
		log.Println(err)
	}

	// Handlers for website & API
	// XXX Auth website may use API (extended) with AJAX
	http.HandleFunc("/", auth)
	http.HandleFunc("/api/", api)

	// Captchas
	http.Handle("/captcha/",
		captcha.Server(captcha.StdWidth, captcha.StdHeight))

	// Static files
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static/"))))

	if C.SSL {
		log.Println("Launching on https://localhost:" + C.Port)
		log.Fatal(http.ListenAndServeTLS(":"+C.Port, C.Certificate, C.PKey, nil))
	} else {
		log.Println("Launching on http://localhost:" + C.Port)
		log.Fatal(http.ListenAndServe(":"+C.Port, nil))
	}
}
