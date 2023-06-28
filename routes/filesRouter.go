package routes

import (
	"log"
	"net/http"
	"path"
	"posts/globals"
)

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	session, err := globals.LoginCookie.Get(r, "login")
	if err != nil {
		log.Println("Error getting session")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if session.Values["authenticated"] != true {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	http.ServeFile(w, r, path.Join("public", "index.html"))
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := globals.LoginCookie.Get(r, "login")

	if session.Values["authenticated"] == true {
		http.Redirect(w, r, "/media", http.StatusFound)
		return
	}

	http.ServeFile(w, r, path.Join("public", "auth", "signup.html"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := globals.LoginCookie.Get(r, "login")

	if session.Values["authenticated"] == true {
		http.Redirect(w, r, "/media", http.StatusFound)
		return
	}

	http.ServeFile(w, r, path.Join("public", "auth", "login.html"))
}
