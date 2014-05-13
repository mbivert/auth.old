package main

import (
	"strings"
	"net/smtp"
	"log"
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

	if err := smtp.SendMail("smtp.gmail.com:587", auth, from,
			[]string{to},[]byte(body)); err != nil {
		return MkIErr(err)
	}

	return nil
}

func sendToken(email string, token *Token) error {
	s := db.GetService2(token.Key)

	err := sendEmail(email, "Token for "+s.Name,
		"Hi there,\r\n"+
		"Here is your token for "+s.Name+" ("+s.Url+")"+": "+token.Token)

	if err != nil { log.Println(err); return SMTPErr }

	return nil
}

func checkName(name string) error {
	switch {
	case name == "":							return NoNameErr
	case len(name) >= LenToken:					return LongNameErr
	case strings.Contains(name, "@ \t\n\r"):	return NameFmtErr
	}

	return nil
}

func checkEmail(email string) error {
	switch {
	case email == "":						return NoEmailErr
	case len(email) >= LenToken:			return LongEmailErr
	case !strings.Contains(email, "@"):		return EmailFmtErr
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
		log.Println(err)
		return WrongUser
	}

	return sendToken(email, NewToken(u.Id, Auth.Key))
}

func Login(login string) (string, error) {
	if isToken(login) {
		ntoken := UpdateToken(login)
		if ntoken == "" {
			return "", NoSuchTErr
		}
		return ntoken, nil
	}

	u, err := db.GetUser2(login)
	if err != nil {
		log.Println(err)
		return "", NoSuchErr
	}

	return "", sendToken(u.Email, NewToken(u.Id, Auth.Key))
}

/*func LoginPassword() {
}*/

func Logout(token string) {
	RemoveToken(token)
}

/*func Unregister() {
}*/

func IsAdmin(token string) bool {
	return db.IsAdmin(OwnerToken(token))
}

func AddService(name, url, address, email string) (string, error) {
	if name == "" || url == "" {
		return "", EmptyFieldsErr
	}

	if ServiceMode == Disabled { return "ko", nil }

	s := Service{ -1, name, url, randomString(64), false, address, email }
	if err := db.AddService(&s); err != nil {
		return "", err
	}

	if ServiceMode == Automatic {
		db.SetMode(s.Id, true)
		return s.Key, nil
	}

	// Manual
	SendAdmin("New Service "+s.Name,
			"Hi there,\r\n"+
			s.Name + " ("+s.Address+", "+s.Url+") asks for landing.")

	return "ok", nil

}

func CheckService(key, address string) bool {
	s := db.GetService2(key)
	if s == nil  { return false }

	return s.Address == address && s.Mode
}

func SendAdmin(subject, msg string) {
	if emails, err := db.GetAdminMail(); err != nil {
		log.Println(err)
		return
	} else {
		for _, to := range emails {
			if err := sendEmail(to, subject, msg); err != nil {
				log.Println(err)
			}
		}
	}
}
