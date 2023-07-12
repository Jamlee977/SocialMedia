package routes

import (
	"encoding/json"
	"net/http"
	"posts/firebase"
	"posts/globals"
	"posts/models"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type profileDetails struct {
    Name string `json:"name"`
    Id string `json:"id"`
}

func FollowUser(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(r.URL.Path, "/")
    userId := parts[len(parts)-2]
    _, err := uuid.Parse(userId)
    if err != nil {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }

    var wg sync.WaitGroup
    var followerDocumentId string
    var userDocumentId string
    var followerErr error
    var userErr error

    wg.Add(2)
    go func() {
        defer wg.Done()
        var account firebase.AccountRepository = &firebase.Account{}
        session, _ := globals.LoginCookie.Get(r, "login")
        followerId := session.Values["id"].(string)
        followerDocumentId, followerErr = account.GetDocumentIdByUuid(followerId)
    }()
    go func() {
        defer wg.Done()
        var account firebase.AccountRepository = &firebase.Account{}
        userDocumentId, userErr = account.GetDocumentIdByUuid(userId)
    }()
    wg.Wait()

    if followerErr != nil || userErr != nil {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }

    session, _ := globals.LoginCookie.Get(r, "login")
    if session.Values["id"] == userId {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }

    var account firebase.AccountRepository = &firebase.Account{}
    account.AddFollower(followerDocumentId, userDocumentId)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func UnfollowUser(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(r.URL.Path, "/")
    userId := parts[len(parts)-2]
    _, err := uuid.Parse(userId)
    if err != nil {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }

    var wg sync.WaitGroup
    var userDocumentId string
    var followerDocumentId string
    var userErr error
    var followerErr error

    wg.Add(2)
    go func() {
        defer wg.Done()
        var account firebase.AccountRepository = &firebase.Account{}
        userDocumentId, userErr = account.GetDocumentIdByUuid(userId)
    }()
    go func() {
        defer wg.Done()
        session, _ := globals.LoginCookie.Get(r, "login")
        followerId := session.Values["id"].(string)
        var account firebase.AccountRepository = &firebase.Account{}
        followerDocumentId, followerErr = account.GetDocumentIdByUuid(followerId)
    }()
    wg.Wait()

    if userErr != nil || followerErr != nil {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }

    session, _ := globals.LoginCookie.Get(r, "login")
    if session.Values["id"] == userId {
        http.Redirect(w, r, "/media", http.StatusNotFound)
        return
    }

    var account firebase.AccountRepository = &firebase.Account{}
    account.RemoveFollower(followerDocumentId, userDocumentId)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
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

func GetProfileDetailsOnMediaPage(w http.ResponseWriter, r *http.Request) {
    var user profileDetails

    session, err := globals.LoginCookie.Get(r, "login")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    firstName := session.Values["firstName"].(string)
    lastName := session.Values["lastName"].(string)
    user.Name = firstName + " " + lastName
    id := session.Values["id"].(string)
    user.Id = id

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
	    http.Redirect(w, r, "/signup", http.StatusBadRequest)
        return
	}

	if len(user.Password) < 8 {
        http.Redirect(w, r, "/signup", http.StatusBadRequest)
		return
	}

	var accountRepository firebase.AccountRepository = &firebase.Account{}

	err = accountRepository.CreateAccount(&user)
	if err != nil {
        http.Redirect(w, r, "/signup", http.StatusBadRequest)
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
        http.Redirect(w, r, "/login", http.StatusBadRequest)
        return
	}

	if user == nil {
        http.Redirect(w, r, "/login", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusBadRequest)
        return
	}

	session, err := globals.LoginCookie.Get(r, "login")
	if err != nil {
        http.Redirect(w, r, "/login", http.StatusInternalServerError)
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
        http.Redirect(w, r, "/login", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := globals.LoginCookie.Get(r, "login")
	if err != nil {
        http.Redirect(w, r, "/login", http.StatusInternalServerError)
		return
	}

	session.Options.MaxAge = -1
	session.Values["authenticated"] = false

	err = session.Save(r, w)
	if err != nil {
        http.Redirect(w, r, "/login", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
