package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	port       = flag.String("port", "8081", "Listening HTTP port")
	ssl        = flag.Bool("ssl", true, "Use SSL")
	AuthServer = flag.String("aas", "https://localhost:8080/", "AAS to use")
	AuthClient = &http.Client{}
	Key        = flag.String("key", "nJTkE9XjmBTM29M4riZWR7Zy9iuaB4EkzJYbqmq3DfyDSXeaU9qBv9mme6NEiaji", "Service's key for the AAS")
	DataDir    = flag.String("data", "./data/", "Data directory")
)

func ko(w http.ResponseWriter) {
	http.Error(w, "ko", http.StatusBadRequest)
}

func ok(w http.ResponseWriter) {
	w.Write([]byte("ok"))
}

// Helper to contact API
func mkr(descr string) string {
	r, err := AuthClient.Get(*AuthServer + "/api/" + descr + "&key=" + *Key)
	if err != nil {
		// XXX watch out, err may contain sensible data (key)
		log.Println(err)
		return "ko"
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil && err != io.EOF {
		log.Println(err)
		return "ko"
	}

	return string(body)
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
	mkr("logout?token="+token)
}

// HTTP
const indexcontent = `
<!DOCTYPE html>
<html>
	<head>
		<title>Storing test server for AAS</title>
	</head>
	<body>
		<p> Use through API only. </p>
		<p>
			For instance, the <a href="https://localhost:8082">example server</a>
			use it to store user data.
		</p>
	</body>
</html>
`

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(indexcontent))
}

func get(w http.ResponseWriter, r *http.Request, ids string) {
	if data, err := ioutil.ReadFile(*DataDir + "/" + ids); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
	} else {
		w.Write(data)
	}
}

func store(w http.ResponseWriter, r *http.Request, ids string) {
	data := []byte(r.FormValue("data"))
	if err := ioutil.WriteFile(*DataDir+"/"+ids, data, 0600); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
	} else {
		ok(w)
	}
}

func api(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	ad := info(token)
	if ad == nil {
		log.Println("bad token: ", token)
		ko(w)
		return
	}

	ids := strconv.FormatInt(int64(ad.Uid), 10)

	switch r.URL.Path[5:] {
	case "store":
		store(w, r, ids)
	case "get":
		get(w, r, ids)
	}
	logout(token)
}

func loadAuthCert() {
	pem, err := ioutil.ReadFile("auth-cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	certs := x509.NewCertPool()
	if !certs.AppendCertsFromPEM(pem) {
		log.Fatal(errors.New("can't add auth-cert.pem"))
	}
	AuthClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
	}
}

func main() {
	flag.Parse()
	loadAuthCert()

	if err := os.Mkdir(*DataDir, 0700); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/api/", api)

	if *ssl {
		log.Print("Launching on https://localhost:" + *port)
		log.Fatal(http.ListenAndServeTLS(":"+*port, "cert.pem", "key.pem", nil))
	} else {
		log.Print("Launching on http://localhost:" + *port)
		log.Fatal(http.ListenAndServe(":"+*port, nil))
	}
}
