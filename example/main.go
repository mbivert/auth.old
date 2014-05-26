package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"github.com/gorilla/securecookie"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct{ URL, Key string }

var (
	port   = flag.String("port", "8082", "Listening HTTP port")
	ssl    = flag.Bool("ssl", true, "Use SSL")
	AStore = &Store{"https://localhost:8081/", "storexample"}
	confd  = "./conf/"
	Conf   = map[string]Config{}
	Client = &http.Client{}
)

// Configuration
func loadConfig(f string, certs *x509.CertPool) error {
	name := path.Base(strings.TrimSuffix(f, ".conf"))

	var c Config

	content, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	for n, line := range bytes.Split(content, []byte("\n")) {
		vals := bytes.Split(line, []byte("="))
		// silently ignore bad lines.
		if len(vals) != 2 {
			continue
		}

		switch string(vals[0]) {
		case "url":
			c.URL = string(vals[1])
		case "key":
			c.Key = string(vals[1])
		case "cert":
			pem, err := ioutil.ReadFile(confd + "/" + string(vals[1]))
			if err != nil {
				return err
			}
			if !certs.AppendCertsFromPEM(pem) {
				return errors.New(f + ":" + strconv.Itoa(n) + "can't add certificate")
			}
		default:
			log.Println("Unknown field " + string(vals[0]) + " (=" + string(vals[1]) + ")")
		}
	}

	if c.URL == "" {
		return errors.New(f + ": missing url=")
	}
	if c.Key == "" {
		return errors.New(f + ": missing key=")
	}

	Conf[name] = c

	return nil
}

func loadConfigs() {
	// XXX in real life, worth keeping root certificates
	// for auth server with real x509.
	certs := x509.NewCertPool()
	// load auth servers
	fs, err := filepath.Glob(confd + "/*.conf")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fs {
		if err := loadConfig(f, certs); err != nil {
			log.Fatal(err)
		}
	}

	// create client
	Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
	}
}

// Auth API
func mkr(a, descr string) string {
	resp, err := Client.Get(Conf[a].URL + "/api/" + descr + "&key=" + Conf[a].Key)
	if err != nil {
		// XXX watch out, err may contain sensible data (key)
		log.Println(err)
		return "ko"
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "ko"
	}

	return string(body)
}

func login(a, login string) string {
	return mkr(a, "login?login="+login)
}

func chain(a, token string) string {
	return mkr(a, "chain?token="+token)
}

type AuthData struct {
	Uid    int32
	Name   string
	Email  string
	Server string
	Value  string
}

func info(a, token string) *AuthData {
	res := strings.Split(mkr(a, "info?token="+token), "\n")

	if uid, err := strconv.ParseInt(res[0], 10, 32); err == nil {
		return &AuthData{int32(uid), res[1], res[2], a, ""}
	}

	// ko
	return nil
}

func logout(a, token string) {
	mkr(a, "logout?token="+token)
}

func bridge(a, name, token string) string {
	return mkr(a, "bridge?token="+token+"&name="+name)
}

// Store API
type Store struct {
	Url  string
	Name string
}

func (s *Store) Put(token, data string) error {
	v := url.Values{"token": {token}, "data": {data}}
	r, err := Client.Get(s.Url + "/api/store?" + v.Encode())
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if string(body) == "ko" {
		return errors.New("Cannot store data")
	}

	return nil
}

func (s *Store) Get(token string) (string, error) {
	r, err := Client.Get(s.Url + "/api/get?token=" + token)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Cookie management
var (
	hashKey  = []byte(securecookie.GenerateRandomKey(32))
	blockKey = []byte(securecookie.GenerateRandomKey(32))
	s        = securecookie.New(hashKey, blockKey)
	ctoken   = "test-token"
)

func setToken(w http.ResponseWriter, token string, ad *AuthData) error {
	value := map[string]interface{}{
		"token":  token,
		"uid":    ad.Uid,
		"name":   ad.Name,
		"email":  ad.Email,
		"server": ad.Server,
	}

	if encoded, err := s.Encode(ctoken, value); err == nil {
		cookie := &http.Cookie{
			Name:  ctoken,
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	} else {
		return err
	}

	return nil
}

func getToken(r *http.Request) (string, AuthData, error) {
	cookie, err := r.Cookie(ctoken)
	if err != nil {
		return "", AuthData{}, err
	}

	v := map[string]interface{}{}
	err = s.Decode(ctoken, cookie.Value, &v)

	if err != nil {
		return "", AuthData{}, err
	}

	return v["token"].(string),
		AuthData{v["uid"].(int32),
			v["name"].(string),
			v["email"].(string),
			v["server"].(string),
			"",
		}, nil
}

func unsetToken(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   ctoken,
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

// HTTP
const srcLogin = `
<!DOCTYPE html>
<html>
	<head>
		<title>Test server for AAS</title>
	</head>
	<body>
		<form action="/" method="post">
			<select name="server">
			{{ range $i, $v := . }}
				<option value="{{ $i }}"> {{ $i }} ({{ $v.URL }}) </option>
			{{ end }}
			</select>
			<input autocomplete="off" name="login"
				type="text" class="form-control"
				placeholder="Token, username or email" />
			<p>
				How does it work (assuming you have an AAS account already)
				<ol>
					<li>If you have no token, enter either name or email</li>
					<li>A token will be generated for this service/user</li>
					<li>Go to your sessions page on AAS and get the token</li>
					<li>Enter it above and... you're done!</li>
					</ul>
				</ol>
			</p>
			<input name="login" type="submit" value="Login" />
		</form>
	</body>
</html>
`
const srcUser = `
<!DOCTYPE html>
<html>
	<head>
		<title>Test server for AAS</title>
	</head>
	<body>
		<p>Connected.</p>
		<p>User data from auth server:</p>
		<ul>
			<li>UID : {{ .Uid }}</li>
			<li>Name : {{ .Name }}</li>
			<li>Email : {{ .Email }}</li>
		</ul>
		<p> User data from the store: </p>
		<form action="/user" method="post">
			<textarea name="data">{{ .Value }}</textarea>
			<input type="submit" value="Store new data" />
		</form>
		<p>
			When <a href="/leave">leaving</a>, the session will
			have also disappear from your AAS.
		</p>
	</body>
</html>
`

var tmplLogin = template.Must(template.New("form").Parse(srcLogin))
var tmplUser = template.Must(template.New("user").Parse(srcUser))

func index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if err := tmplLogin.Execute(w, &Conf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		usr, srv := r.FormValue("login"), r.FormValue("server")

		res := login(srv, usr)
		if res == "ko" {
			http.Error(w, "ko", http.StatusInternalServerError)
		} else if res == "ok" {
			if err := setToken(w, usr, info(srv, usr)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				http.Redirect(w, r, "/user", http.StatusFound)
			}
		} else /* res == "new" */ {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
}

func user(w http.ResponseWriter, r *http.Request) {
	token, ad, err := getToken(r)
	if err != nil {
		http.Error(w, "Not connected", http.StatusFound)
		return
	}
	token = chain(ad.Server, token)
	if token != "ko" {
		setToken(w, token, &ad)
	} else {
		http.Error(w, "Bad token", http.StatusFound)
	}

	btoken := bridge(ad.Server, AStore.Name, token)
	if btoken == "ko" {
		http.Error(w, "cannot bridge", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if err := AStore.Put(btoken, r.FormValue("data")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if ad.Value, err = AStore.Get(btoken); err != nil {
		ad.Value = "Error while storing: " + err.Error()
	}
	if err := tmplUser.Execute(w, &ad); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	logout(ad.Server, btoken)
}

func leave(w http.ResponseWriter, r *http.Request) {
	if token, ad, err := getToken(r); err == nil {
		logout(ad.Server, token)
		unsetToken(w)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func main() {
	flag.Parse()
	loadConfigs()

	http.HandleFunc("/", index)
	http.HandleFunc("/user", user)
	http.HandleFunc("/leave", leave)

	if *ssl {
		log.Print("Launching on https://localhost:" + *port)
		log.Fatal(http.ListenAndServeTLS(":"+*port, "cert.pem", "key.pem", nil))
	} else {
		log.Print("Launching on http://localhost:" + *port)
		log.Fatal(http.ListenAndServe(":"+*port, nil))
	}
}
