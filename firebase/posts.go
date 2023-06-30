package firebase

import (
	"context"
	"log"
	"posts/globals"
	"posts/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type PostsRepository interface {
	AddPost(post *models.Post, author string, authorId string) error
	GetPosts() ([]*models.Post, error)
    GetPostByAuthorId(authorId string) ([]models.Post, error)
}

type Posts struct{}

func getFirebasePostsClient(ctx context.Context) (*firestore.Client, error) {
    opt := option.WithCredentialsJSON([]byte(globals.ServiceAccountKey))
	client, err := firestore.NewClient(ctx, globals.ProjectId, opt)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}

	return client, nil
}

func (*Posts) AddPost(post *models.Post, author string, authorId string) error {
	ctx := context.Background()
	client, err := getFirebasePostsClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return err
	}
	defer client.Close()

	post.Author = author
    post.AuthorId = authorId

	_, _, err = client.Collection(globals.PostsCollectionName).Add(ctx, map[string]string{
		"Author":  post.Author,
		"Content": post.Content,
        "AuthorId": post.AuthorId,
	})

	if err != nil {
		log.Fatalf("Failed to add post: %v", err)
		return err
	}

	return nil
}

func (*Posts) GetPosts() ([]*models.Post, error) {
	ctx := context.Background()
	client, err := getFirebasePostsClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}
	defer client.Close()

	var posts []*models.Post
	iter := client.Collection(globals.PostsCollectionName).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
			return nil, err
		}

		var post models.Post
		doc.DataTo(&post)
		posts = append(posts, &post)
	}

	return posts, nil
}

func (*Posts) GetPostByAuthorId(id string) ([]models.Post, error) {
    ctx := context.Background()
	client, err := getFirebasePostsClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}
	defer client.Close()

    var posts []models.Post
    iter := client.Collection(globals.PostsCollectionName).Where("AuthorId", "==", id).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
			return nil, err
		}

		var post models.Post
		doc.DataTo(&post)
        posts = append(posts, post)
	}

	return posts, nil
}
