package firebase

import (
	"context"
	"fmt"
	"log"
	"posts/globals"
	"posts/models"

	"cloud.google.com/go/firestore"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
)

type AccountRepository interface {
	CreateAccount(user *models.User) error
	FindAccountByEmail(email *string) (*models.User, error)
}

type Account struct{}

func getFirebaseUserClient(ctx context.Context) (*firestore.Client, error) {
	opt := option.WithCredentialsFile(globals.ServiceAccountKeyPath)
	client, err := firestore.NewClient(ctx, globals.ProjectId, opt)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}

	return client, nil
}

func (*Account) CreateAccount(user *models.User) error {
	ctx := context.Background()
	client, err := getFirebaseUserClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return err
	}
	defer client.Close()

	query := client.Collection(globals.UsersCollectionName).Where("Email", "==", user.Email).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
		return err
	}

	if len(docs) > 0 {
		return fmt.Errorf("User already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
		return err
	}

	_, _, err = client.Collection(globals.UsersCollectionName).Add(ctx, map[string]string{
		"Email":     user.Email,
		"Password":  string(hashedPassword),
		"FirstName": user.FirstName,
		"LastName":  user.LastName,
	})

	if err != nil {
		log.Fatalf("Failed to add user: %v", err)
		return err
	}

	return nil
}

func (*Account) FindAccountByEmail(email *string) (*models.User, error) {
	ctx := context.Background()
	client, err := getFirebaseUserClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}
	defer client.Close()

	query := client.Collection(globals.UsersCollectionName).Where("Email", "==", *email).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
		return nil, err
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("User not found")
	}

	var user models.User
	for _, doc := range docs {
		doc.DataTo(&user)
	}

	return &user, nil
}
