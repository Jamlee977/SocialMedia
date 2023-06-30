package routes

import (
	"encoding/json"
	"net/http"
	"posts/firebase"
	"posts/globals"
	"posts/models"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type profileDetails struct {
    Name string `json:"name"`
}

func GetProfilePosts(w http.ResponseWriter, r *http.Request) {
    var postsRepository firebase.PostsRepository = &firebase.Posts{}

    vars := mux.Vars(r)
    authorId := vars["userId"]

    if strings.TrimSpace(authorId) == "" {
        http.Redirect(w, r, "/media", http.StatusSeeOther)
        return
    }
    
    var account firebase.AccountRepository = &firebase.Account{}
    user, err := account.FindAccountByUuid(authorId)
    if err != nil {
        http.Redirect(w, r, "/media", http.StatusSeeOther)
        return
    }

    posts, err := postsRepository.GetPostByAuthorId(user.Id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func GetProfileDetails(w http.ResponseWriter, r *http.Request) {
    var user profileDetails

    session, err := globals.LoginCookie.Get(r, "login")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    firstName := session.Values["firstName"].(string)
    lastName := session.Values["lastName"].(string)
    user.Name = firstName + " " + lastName

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func AddPost(w http.ResponseWriter, r *http.Request) {
	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, err := globals.LoginCookie.Get(r, "login")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	firstName := session.Values["firstName"].(string)
	lastName := session.Values["lastName"].(string)
	post.Author = firstName + " " + lastName

    userId := session.Values["id"].(string)
    post.AuthorId = userId

	var postsRepository firebase.PostsRepository = &firebase.Posts{}
	err = postsRepository.AddPost(&post, post.Author, post.AuthorId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	var postsRepository firebase.PostsRepository = &firebase.Posts{}
	posts, err := postsRepository.GetPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func SignupAfterCheckingTheDatabase(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.Email = r.FormValue("email")
	user.Password = r.FormValue("password")
	user.FirstName = r.FormValue("first_name")
	user.LastName = r.FormValue("last_name")
	confirmPassword := r.FormValue("confirm_password")

	if user.Password != confirmPassword {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Passwords do not match"))
		return
	}

	if len(user.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Password must be longer than 8 characters"))
		return
	}

	var accountRepository firebase.AccountRepository = &firebase.Account{}

	err = accountRepository.CreateAccount(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginAfterCheckingTheDatabase(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var accountRepository firebase.AccountRepository = &firebase.Account{}
	user, err := accountRepository.FindAccountByEmail(&email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, err := globals.LoginCookie.Get(r, "login")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    session.Values["id"] = user.Id
	session.Values["email"] = user.Email
	session.Values["firstName"] = user.FirstName
	session.Values["lastName"] = user.LastName
	session.Values["loginTime"] = time.Now().Unix()
	session.Values["authenticated"] = true

	session.Options.MaxAge = 60 * 60 * 24 * 7
	session.Options.Secure = true
	session.Options.SameSite = http.SameSiteStrictMode
	session.Options.HttpOnly = true

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := globals.LoginCookie.Get(r, "login")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Options.MaxAge = -1
	session.Values["authenticated"] = false

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
