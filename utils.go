package main

import (
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
)

const (
	alnum 			=	"abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789"
)

// Generate random string of n bytes
func randomString(n int) string {
	buf := make([]byte, n)

	for i := 0; i < C.LenToken; i++ {
		buf[i] = alnum[rand.Intn(len(alnum))]
	}

	return string(buf)
}

// WriteFiles write the files it's given as argument to w
func writeFiles(w http.ResponseWriter, files ...string) error {
	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		w.Write(b)
	}
	return nil
}

func ok(w http.ResponseWriter) {
	w.Write([]byte("ok"))
}
func ko(w http.ResponseWriter) {
	w.Write([]byte("ko"))
}

var hashKey = []byte(securecookie.GenerateRandomKey(32))
var blockKey = []byte(securecookie.GenerateRandomKey(32))
var s = securecookie.New(hashKey, blockKey)

func SetToken(w http.ResponseWriter, token string) error {
	encoded, err := s.Encode("auth-token", token)
	if err != nil { return MkIErr(err) }

	cookie := &http.Cookie {
		Name	:	"auth-token",
		Value	:	encoded,
		Path	:	"/",
	}
	http.SetCookie(w, cookie)

	return nil
}

func UnsetToken(w http.ResponseWriter) {
	cookie := &http.Cookie {
		Name	:	"auth-token",
		Value	:	"",
		Path	:	"/",
		MaxAge	:	-1,
	}
	http.SetCookie(w, cookie)
}


func VerifyToken(r *http.Request) (token string, err error) {
	cookie, err := r.Cookie("auth-token")
	if err == nil { 
		err = s.Decode("auth-token", cookie.Value, &token)
	}

	if err != nil || !CheckToken(token) {
		return "", MouldyCookie
	}

	return token, nil
}

func SetInfo(w http.ResponseWriter, msg string) {
	cookie := &http.Cookie {
		Name	:	"auth-info",
		Value	:	strings.Replace(msg, " ", "_", -1),
		Path	:	"/",
	}
	http.SetCookie(w, cookie)
}

func SetError(w http.ResponseWriter, err error) {
	SetInfo(w, "Error: "+err.Error())
}
