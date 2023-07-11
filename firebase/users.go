package firebase

import (
	"context"
	"fmt"
	"log"
	"posts/globals"
	"posts/models"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type AccountRepository interface {
	CreateAccount(user *models.User) error
	FindAccountByEmail(email *string) (*models.User, error)
    FindAccountByUuid(id string) (*models.User, error)
    AddFollower(followerId string, followingId string) error
    GetDocumentIdByUuid(uuid string) (string, error)
    IsFollowing(firstUuid string, secondUuid string) (bool, error)
}

type Account struct{}

func getFirebaseUserClient(ctx context.Context) (*firestore.Client, error) {
    opt := option.WithCredentialsJSON([]byte(globals.ServiceAccountKey))
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

    user.Id = uuid.New().String()

    followers := make([]string, 0)
    following := make([]string, 0)

    _, _, err = client.Collection(globals.UsersCollectionName).Add(ctx, map[string]interface{}{
        "Id":        user.Id,
        "Email":     user.Email,
        "Password":  string(hashedPassword),
        "FirstName": user.FirstName,
        "LastName":  user.LastName,
        "Followers": followers,
        "Following": following,
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

func (*Account) FindAccountByUuid(id string) (*models.User, error) {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return nil, err
    }
    defer client.Close()

    query := client.Collection(globals.UsersCollectionName).Where("Id", "==", id).Limit(1)
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

func (*Account) GetDocumentIdByUuid(uuid string) (string, error) {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return "", err
    }
    defer client.Close()

    collection := client.Collection(globals.UsersCollectionName)
    query := collection.Query

    documentIterator := query.Documents(ctx)
    for {
        doc, err := documentIterator.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            log.Fatalf("Failed to iterate: %v", err)
            return "", err
        }
        var user models.User
        doc.DataTo(&user)
        if user.Id == uuid {
            return doc.Ref.ID, nil
        }
    }

    return "", fmt.Errorf("User not found")
}

func (*Account) AddFollower(followerId string, followingId string) error {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return err
    }
    defer client.Close()

    followingRef := client.Collection(globals.UsersCollectionName).Doc(followingId)
    _, err = followingRef.Update(ctx, []firestore.Update{
        {
            Path:  "Followers",
            Value: firestore.ArrayUnion(followerId),
        },
    })

    if err != nil {
        log.Fatalf("Failed adding follower: %v", err)
        return err
    }

    followerRef := client.Collection(globals.UsersCollectionName).Doc(followerId)
    _, err = followerRef.Update(ctx, []firestore.Update{
        {
            Path:  "Following",
            Value: firestore.ArrayUnion(followingId),
        },
    })

    if err != nil {
        log.Fatalf("Failed adding following: %v", err)
        return err
    }

    return nil
}

func (*Account) IsFollowing(firstUuid string, secondUuid string) (bool, error) {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return false, err
    }
    defer client.Close()

    firstId, err := getDocumentIdByUuid(firstUuid)
    if err != nil {
        log.Fatalf("Failed to get first id: %v", err)
        return false, err
    }

    secondId, err := getDocumentIdByUuid(secondUuid)
    if err != nil {
        log.Fatalf("Failed to get second id: %v", err)
        return false, err
    }

    firstRef := client.Collection(globals.UsersCollectionName).Doc(firstId)

    // is first following second
    firstSnapshot, err := firstRef.Get(ctx)
    if err != nil {
        log.Fatalf("Failed to get first snapshot: %v", err)
        return false, err
    }

    var firstAccount models.User
    err = firstSnapshot.DataTo(&firstAccount)
    if err != nil {
        log.Fatalf("Failed to get first account: %v", err)
        return false, err
    }

    for _, following := range firstAccount.Following {
        if following == secondId {
            return true, nil
        }
    }

    return false, nil
}

func getDocumentIdByUuid(uuid string) (string, error) {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return "", err
    }
    defer client.Close()

    iter := client.Collection(globals.UsersCollectionName).Documents(ctx)
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            log.Fatalf("Failed to iterate: %v", err)
            return "", err
        }

        var user models.User
        doc.DataTo(&user)
        if user.Id == uuid {
            return doc.Ref.ID, nil
        }
    }

    return "", fmt.Errorf("User not found")
}
