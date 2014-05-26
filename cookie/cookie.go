package cookie

import (
	"github.com/gorilla/securecookie"
	"net/http"
)

var (
	hashKey  = []byte(securecookie.GenerateRandomKey(32))
	blockKey = []byte(securecookie.GenerateRandomKey(32))
	s        = securecookie.New(hashKey, blockKey)
)

func Set(w http.ResponseWriter, name, value string) error {
	encoded, err := s.Encode(name, value)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:  name,
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, cookie)

	return nil
}

func Get(r *http.Request, name string) (string, error) {
	var value string

	cookie, err := r.Cookie(name)
	if err == nil {
		err = s.Decode(name, cookie.Value, &value)
	}
	return value, err
}

func Unset(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}
