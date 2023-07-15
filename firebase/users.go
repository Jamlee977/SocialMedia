package firebase

import (
	"context"
	"fmt"
	"log"
	"posts/globals"
	"posts/models"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"google.golang.org/api/option"
)

type AccountRepository interface {
	CreateAccount(user *models.User) error
	FindAccountByEmail(email *string) (*models.User, error)
    FindAccountByUuid(id string) (*models.User, error)
    AddFollower(followerId string, followingId string) error
    RemoveFollower(followerId string, followingId string) error
    GetDocumentIdByUuid(uuid string) (string, error)
    IsFollowing(firstUuid string, secondUuid string) (bool, error)
    UpdateEmail(docId string, email string) error
    UpdateFirstName(docId string, firstName string) error
    UpdateLastName(docId string, lastName string) error
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

var userCache = make(map[string]*models.User)

func (*Account) FindAccountByUuid(id string) (*models.User, error) {
	if user, ok := userCache[id]; ok {
		return user, nil
	}

	ctx := context.Background()
	client, err := getFirebaseUserClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}
	defer client.Close()

	query := client.Collection(globals.UsersCollectionName).Where("Id", "==", id).Limit(1)
	snapshots, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("User not found")
	}

	var user models.User
	if err := snapshots[0].DataTo(&user); err != nil {
		return nil, err
	}

	userCache[id] = &user

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
    query := collection.Where("Id", "==", uuid).Limit(1)

    docs, err := query.Documents(ctx).GetAll()
    if err != nil {
        log.Fatalf("Failed to get user: %v", err)
        return "", err
    }

    if len(docs) == 0 {
        return "", fmt.Errorf("User not found")
    }

    return docs[0].Ref.ID, nil
}

func (*Account) AddFollower(followerId string, followingId string) error {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        return fmt.Errorf("failed to create client: %v", err)
    }
    defer client.Close()

    followingRef := client.Collection(globals.UsersCollectionName).Doc(followingId)
    followerRef := client.Collection(globals.UsersCollectionName).Doc(followerId)

    var wg sync.WaitGroup
    var updateErr error

    wg.Add(2)
    go func() {
        defer wg.Done()
        _, err := followingRef.Update(ctx, []firestore.Update{
            {
                Path:  "Followers",
                Value: firestore.ArrayUnion(followerId),
            },
        })
        if err != nil {
            updateErr = fmt.Errorf("failed adding follower: %v", err)
        }
    }()

    go func() {
        defer wg.Done()
        _, err := followerRef.Update(ctx, []firestore.Update{
            {
                Path:  "Following",
                Value: firestore.ArrayUnion(followingId),
            },
        })
        if err != nil {
            updateErr = fmt.Errorf("failed adding following: %v", err)
        }
    }()

    wg.Wait()

    return updateErr
}

func (*Account) RemoveFollower(followerId string, followingId string) error {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return err
    }
    defer client.Close()

    followingRef := client.Collection(globals.UsersCollectionName).Doc(followingId)
    followerRef := client.Collection(globals.UsersCollectionName).Doc(followerId)

    var wg sync.WaitGroup
    var updateErr error

    wg.Add(2)

    go func() {
        defer wg.Done()
        _, err := followingRef.Update(ctx, []firestore.Update{
            {
                Path:  "Followers",
                Value: firestore.ArrayRemove(followerId),
            },
        })

        if err != nil {
            log.Fatalf("Failed removing follower: %v", err)
            updateErr = err
        }
    }()

    go func() {
        defer wg.Done()
        _, err := followerRef.Update(ctx, []firestore.Update{
            {
                Path:  "Following",
                Value: firestore.ArrayRemove(followingId),
            },
        })

        if err != nil {
            log.Fatalf("Failed removing following: %v", err)
            updateErr = err
        }
    }()

    wg.Wait()

    return updateErr
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
        return false, err
    }

    secondId, err := getDocumentIdByUuid(secondUuid)
    if err != nil {
        return false, err
    }

    followingSnapshot, err := client.Collection(globals.UsersCollectionName).Doc(firstId).Get(ctx)
    if err != nil {
        log.Fatalf("Failed to get following: %v", err)
        return false, err
    }

    var followingData struct {
        Following []string
    }
    if err := followingSnapshot.DataTo(&followingData); err != nil {
        log.Fatalf("Failed to parse following data: %v", err)
        return false, err
    }

    for _, id := range followingData.Following {
        if id == secondId {
            return true, nil
        }
    }

    return false, nil
}

func (*Account) UpdateEmail(docId, email string) error {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return err
    }
    defer client.Close()

    accountRef := client.Collection(globals.UsersCollectionName).Doc(docId)

    _, err = accountRef.Update(ctx, []firestore.Update{
        {Path: "Email", Value: email},
    })

    if err != nil {
        log.Fatalf("Failed updating user: %v", err)
        return err
    }

    return nil
}

func (*Account) UpdateFirstName(docId, firstName string) error {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return err
    }
    defer client.Close()

    accountRef := client.Collection(globals.UsersCollectionName).Doc(docId)

    _, err = accountRef.Update(ctx, []firestore.Update{
        {Path: "FirstName", Value: firstName},
    })

    if err != nil {
        log.Fatalf("Failed updating user: %v", err)
        return err
    }

    return nil
}

func (*Account) UpdateLastName(docId, lastName string) error {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return err
    }
    defer client.Close()

    accountRef := client.Collection(globals.UsersCollectionName).Doc(docId)

    _, err = accountRef.Update(ctx, []firestore.Update{
        {Path: "LastName", Value: lastName},
    })

    if err != nil {
        log.Fatalf("Failed updating user: %v", err)
        return err
    }

    return nil
}

func getDocumentIdByUuid(uuid string) (string, error) {
    ctx := context.Background()
    client, err := getFirebaseUserClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
        return "", err
    }
    defer client.Close()

    collection := client.Collection(globals.UsersCollectionName)
    query := collection.Where("Id", "==", uuid).Limit(1)

    docs, err := query.Documents(ctx).GetAll()
    if err != nil {
        log.Fatalf("Failed to get user: %v", err)
        return "", err
    }

    if len(docs) == 0 {
        return "", fmt.Errorf("User not found")
    }

    return docs[0].Ref.ID, nil
}
