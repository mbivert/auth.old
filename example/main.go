package main

import (
	"flag"
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"html/template"
)

var port = flag.String("port", "8082", "Listening HTTP port")

// -- Configuration
const (
	authserver = "http://localhost:8080/"
	key        = "gY2kjVxPYQVKMwc9sap1pgfxpRiNucmShUftMCg2bwTtk5SJLyCZqZ4EWwhRWdkT"
)

// Helper to contact API
func mkr(descr string) string {
	resp, err := http.Get(authserver + "/api/" + descr + "&key=" + key)
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

// API
func login(login string) string {
	return mkr("login?login=" + login)
}

func chain(token string) string {
	return mkr("chain?token=" + token)
}

type AuthData struct {
	Uid   int32
	Name  string
	Email string
}

func info(token string) *AuthData {
	res := strings.Split(mkr("info?token="+token), "\n")

	if uid, err := strconv.ParseInt(res[0], 10, 32); err == nil {
		return &AuthData{int32(uid), res[1], res[2]}
	}

	// ko
	return nil
}

func logout(token string) {
	mkr("logout?token=" + token)
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
		"token"	:	token,
		"uid"	:	ad.Uid,
		"name"	:	ad.Name,
		"email"	:	ad.Email,
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
		AuthData{	v["uid"].(int32),
					v["name"].(string),
					v["email"].(string),
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
const srcForm = `
<!DOCTYPE html>
<html>
	<head>
		<title>Test server for AAS</title>
	</head>
	<body>
		<form action="/" method="post">
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
			<input type="submit" value="Login" />
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
		<p>User data:</p>
		<ul>
			<li>UID : {{ .Uid }}</li>
			<li>Name : {{ .Name }}</li>
			<li>Email : {{ .Email }}</li>
		</ul>
		<p>
			When <a href="/leave">leaving</a>, the session will
			have also disappear from your AAS.
		</p>
	</body>
</html>
`
var tmplUser = template.Must(template.New("user").Parse(srcUser))

func connect(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		l := r.FormValue("login")
		res := login(l)
		if res == "ko" {
			http.Error(w, "ko", http.StatusInternalServerError)
		} else if res == "ok" {
			if err := setToken(w, l, info(l)); err != nil {
				log.Println(err)
				http.Error(w, "ko", http.StatusInternalServerError)
				return
			}
			log.Println("Connected.")
		}
		// new
		http.Redirect(w, r, "/", http.StatusFound)
	case "GET":
		if token, ad, err := getToken(r); err != nil || token == "ko" {
			w.Write([]byte(srcForm))
		} else {
			log.Println(token)
			token = chain(token)
			if token != "ko" {
				setToken(w, token, &ad)
				if err := tmplUser.Execute(w, &ad); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			} else {
				http.Error(w, "Chain failed", http.StatusInternalServerError)
			}
		}
	}
}

func leave(w http.ResponseWriter, r *http.Request) {
	token, _, _ := getToken(r)
	logout(token)
	unsetToken(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	http.HandleFunc("/", connect)
	http.HandleFunc("/leave", leave)

	log.Print("Launching on http://localhost:" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
