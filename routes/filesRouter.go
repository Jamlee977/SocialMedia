package routes

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"posts/firebase"
	"posts/globals"
	"strings"

	"github.com/google/uuid"
)

type Username struct {
    Me string
    Name string
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	session, _ := globals.LoginCookie.Get(r, "login")

    if !isUserLoggedIn(w, r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    id, ok := session.Values["id"].(string)
    if !ok {
        log.Println("Error getting id")
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    template := template.Must(template.ParseFiles(path.Join("public", "index.html")))

    err := template.Execute(w, id)
    if err != nil {
        log.Println(err)
    }
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
    if !isUserLoggedIn(w, r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    parts := strings.Split(r.URL.Path, "/")
    if len(parts) != 3 {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }
    userId := parts[2]
    _, err := uuid.Parse(userId)
    if err != nil {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }

    var account firebase.AccountRepository = &firebase.Account{}
    user, err := account.FindAccountByUuid(userId)
    if err != nil {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }

    username := Username{
        Me: user.Id,
        Name: user.FirstName + " " + user.LastName,
    }

    template := template.Must(template.ParseFiles(path.Join("public", "profile.html")))
    err = template.Execute(w, username)
    if err != nil {
        log.Println(err)
    }
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if isUserLoggedIn(w, r) {
		http.Redirect(w, r, "/media", http.StatusFound)
		return
	}

	http.ServeFile(w, r, path.Join("public", "auth", "signup.html"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if isUserLoggedIn(w, r) {
        http.Redirect(w, r, "/media", http.StatusFound)
        return
    }

	http.ServeFile(w, r, path.Join("public", "auth", "login.html"))
}

func isUserLoggedIn(w http.ResponseWriter, r *http.Request) bool {
    session, _ := globals.LoginCookie.Get(r, "login")

    if session.Values["authenticated"] == true {
        return true
    }

    return false
}
