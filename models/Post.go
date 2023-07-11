package models

type Post struct {
    Author string `json:"author"`
    AuthorId string `json:"authorId"`
    Content string `json:"content"`
}
