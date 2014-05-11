package main

import (
	"errors"
	"strings"
	"net/smtp"
)

// sendEmail send an email to a user.
// XXX use several SMTP according to the destination email
// provider to speed things up.
func sendEmail(to, subject, msg string) error {
	from := "auth.newsome@gmail.com"
	passwd := "awesom auth server"

	body := "To: " + to + "\r\nSubject: " +
		subject + "\r\n\r\n" + msg

	auth := smtp.PlainAuth("", from, passwd, "smtp.gmail.com")

	err := smtp.SendMail("smtp.gmail.com:587", auth, from,
		[]string{to},[]byte(body))

	return err
}

func sendToken(email string, token *Token) error {
	s := services[token.Key]

	err := sendEmail(email, "Token for "+s.Name,
		"Hi there,\r\n"+
		"Here is your token for "+s.Name+" ("+s.Url+")"+": "+token.Token)
	return err
}

func checkName(name string) error {
	switch {
	case name == "":
		return errors.New("Name is mandatory")
	case len(name) >= LenToken:
		return errors.New("Name too long")
	case strings.Contains(name, "@"):
		return errors.New("Name is not an email (@ forbidden)")
	}

	return nil
}

func checkEmail(email string) error {
	switch {
	case email == "":
		return errors.New("Email is mandatory")
	case len(email) >= LenToken:
		return errors.New("Email too long")
	case !strings.Contains(email, "@"):
		return errors.New("Wrong email address format")
	}

	return nil
}

// isToken check whether the login is a token or a name/email.
func isToken(login string) bool { return len(login) == LenToken }

// isEmail check whether the login is a name or an email
func isEmail(login string) bool { return strings.Contains(login, "@") }

// Register add a new user to both database and cache.
// If the registration succeeds, a(n activation) token is
// sent to the user.
func Register(name, email string) error {
	if err := checkName(name); err != nil {
		return err
	}
	if err := checkEmail(email); err != nil {
		return err
	}

	u := User{ -1, name, email, false }

	if err := db.AddUser(&u); err != nil {
		return err
	}

	token := NewToken(u.Id, Auth.Key)

	return sendToken(email, token)
}

func Login(login string) (string, error) {
	if isToken(login) {
		ntoken := UpdateToken(login, Auth.Key)
		if ntoken == "" {
			return "", errors.New("Wrong Token")
		}
		return ntoken, nil
	}

	u := db.GetUser2(login)
	if u == nil { return "", errors.New("Wrong name/email") }

	token := NewToken(u.Id, Auth.Key)

	return "", sendToken(u.Email, token)
}

/*func LoginPassword() {
}*/

func Logout(token string) {
	DelToken(token)
}

/*func Unregister() {
}*/

func IsAdmin(token string) bool {
	return db.IsAdmin(tokens[token])
}

func GetTokens(token string) []*Token {
	return utokens[tokens[token]]
}

func AddService(name, url, address, email string) (string, error) {
	if name == "" || url == "" {
		return "", errors.New("")
	}

	s := Service{ -1, name, url, randomString(64), address, email }
	if err := db.AddService(&s); err != nil {
		return "", err
	}

	services[s.Key] = &s

	return s.Key, nil
}

func CheckService(key, address string) bool {
	s := services[key]
	if s == nil { return false }

	return s.Address == address
}
