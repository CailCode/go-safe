package main

import (
	mod "go-safe/models"
	"net/http"
	"text/template"
	"time"
	"os"
)

// check function for nil value in form
func isNil(value string) bool {
	if value != "" {
		return false
	}
	return true
}

// add / delete cookie
func addCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:    name,
		Value:   value,
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
}

// get cookie value
func getCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	data := cookie.Value
	return data
}

// MAIN
func main() {

	// carico la cartella template per i file statici
	staticPage := http.FileServer(http.Dir("template/"))
	http.Handle("/static/", http.StripPrefix("/static/", staticPage))

	// PAGE
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/home", Home)
	http.HandleFunc("/manager", Manager)

	// LISTEN
	port := os.Getenv("PORT")
	http.ListenAndServe(":" + port)
}

// ** LOGIN **
func Login(w http.ResponseWriter, r *http.Request) {

	// carico template
	login, err := template.ParseFiles("template/login.html")
	mod.Try(err)

	// GET Method
	if r.Method != http.MethodPost {
		login.Execute(w, nil)
		return
	}

	// POST Method
	username := r.FormValue("id")
	password := r.FormValue("password")
	badge := r.FormValue("badge")

	if isNil(username) {
		login.Execute(w, struct{ Error string }{"Inserisci l'ID"})
		return
	}
	if isNil(password) {
		login.Execute(w, struct{ Error string }{"Inserisci la Password"})
		return
	}
	if isNil(badge) {
		login.Execute(w, struct{ Error string }{"Inserisci il Badge"})
		return
	}

	// **** DB *** //
	db := mod.ConnectDB()
	defer db.Close()
	u, err := mod.SelectUser(db, username)

	if err != nil {
		login.Execute(w, struct{ Error string }{"ID non Registrato"})
		return
	}

	password = mod.Hash(password)
	if password != u.Password {
		login.Execute(w, struct{ Error string }{"Credenziali Errate"})
		return
	}

	badge = mod.Hash(badge)
	key := mod.NewKey(badge, password)

	addCookie(w, "id", username, 30*time.Minute)
	addCookie(w, "key", key, 30*time.Minute)

	http.Redirect(w, r, "/manager", 301)
}

// ** REGISTER **
func Register(w http.ResponseWriter, r *http.Request) {

	// carico template
	register, err := template.ParseFiles("template/register.html")
	mod.Try(err)

	// GET Method
	if r.Method != http.MethodPost {
		register.Execute(w, nil)
		return
	}

	// POST Method
	// controllo dei dati inseriti nel form
	u := mod.User{
		Username: r.FormValue("id"),
		Password: r.FormValue("password"),
		Email:    r.FormValue("email"),
	}

	// controllo errori
	if isNil(u.Email) {
		register.Execute(w, struct{ Error string }{"Email obbligatoria"})
		return
	}
	if isNil(u.Username) {
		register.Execute(w, struct{ Error string }{"ID obbligatorio"})
		return
	}
	if isNil(u.Password) {
		register.Execute(w, struct{ Error string }{"Password obbligatoria"})
		return
	}
	if len(u.Password) < 8 {
		register.Execute(w, struct{ Error string }{"Password Debole"})
		return
	}

	if isNil(r.FormValue("badge")) {
		register.Execute(w, struct{ Error string }{"Badge obbligatorio"})
		return
	}

	if len(r.FormValue("badge")) < 4 {
		register.Execute(w, struct{ Error string }{"Badge minimo 4 caratteri"})
		return
	}

	if r.FormValue("badge") != r.FormValue("ripeat") {
		register.Execute(w, struct{ Error string }{"Badge Differenti"})
		return
	}

	//---------DB---------//
	db := mod.ConnectDB()
	defer db.Close()
	// controllo se questo username è stato registrato
	if id, _ := mod.SelectUser(db, u.Username); id != nil {
		register.Execute(w, struct{ Error string }{"ID in uso"})
		return
	}
	//------------------//

	//******** KEY *********//
	badge := r.FormValue("badge")
	badge = mod.Hash(badge)

	u.Password = mod.Hash(u.Password)

	key := mod.NewKey(badge, u.Password)
	//*********************//

	// make new user and start session
	mod.NewUser(db, u)
	// set the "session"
	addCookie(w, "key", key, 30*time.Minute)
	addCookie(w, "id", u.Username, 30*time.Minute)

	http.Redirect(w, r, "/manager", 301)
}

// ** HOME **
func Home(w http.ResponseWriter, r *http.Request) {

	// carico template
	home, err := template.ParseFiles("template/home.html")
	mod.Try(err)

	// Logout
	addCookie(w, "id", "", -1)
	addCookie(w, "key", "", -1)

	home.Execute(w, nil)
}

// ** MANAGER **
func Manager(w http.ResponseWriter, r *http.Request) {

	// carico template
	manager, err := template.ParseFiles("template/manager.html")
	mod.Try(err)

	id := getCookie(r, "id")
	key := getCookie(r, "key")
	if id == "" {
		http.Redirect(w, r, "/home", 301)
	}

	db := mod.ConnectDB()
	defer db.Close()

	// POST Method
	if r.Method == http.MethodPost {
		account := r.FormValue("account")
		password := r.FormValue("password")
		mod.NewVault(db, id, key, account, password)
		http.Redirect(w, r, "/manager", 301)
		return
	}

	v := mod.GetVault(db, id, key)
	manager.Execute(w, struct{ Range []mod.Vault }{v})

}
