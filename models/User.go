package models

type User struct {
    Email string
    FirstName string
    LastName string
    Password string
    Id string
    Followers []string
    Following []string
}
