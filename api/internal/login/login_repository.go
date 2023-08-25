package login

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"interviews/internal/auth"
	data "interviews/internal/users"
	"time"
)

var (
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid password")
)

type Repo struct {
	db         *mongo.Client
	collection *mongo.Collection
}

func (r *Repo) Login(email string, password string) (data.User, error) {
	ctx := context.Background()

	var user data.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return data.User{}, err
	}

	ok := validatePassword(password, user.PassHash)
	if !ok {
		//TODO: might need to add a counter here is the login to many times...Then lock.

		return data.User{}, ErrInvalidPassword
	}

	return data.User{
		ID:    user.ID,
		Email: email,
	}, err
}

func (r *Repo) Register(email string, password string) (*data.User, error) {
	ctx := context.Background()

	var user data.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	if user.Email != "" {
		return nil, ErrUserExists
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	newUser := data.User{
		Name:      "",
		Email:     email,
		PassHash:  hash,
		CreatedAt: time.Now(),
		Role:      "",
		Activated: false,
	}

	_, err = r.collection.InsertOne(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (r *Repo) RegisterGoogle(claims auth.GoogleClaims, password string) (*data.User, error) {
	ctx := context.Background()

	var user data.User
	err := r.collection.FindOne(ctx, bson.M{"email": claims.Email}).Decode(&user)
	if err == nil {
		return nil, ErrUserExists
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	newUser := data.User{
		Name:      claims.Name,
		Email:     claims.Email,
		PassHash:  hash,
		CreatedAt: time.Now(),
		Role:      "",
		Activated: false,
	}

	_, err = r.collection.InsertOne(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func hashPassword(password string) (string, error) {
	// Generate a salt for the bcrypt hash
	salt, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(salt), nil
}

func validatePassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func NewLoginRepository(client *mongo.Client, collection *mongo.Collection) *Repo {
	return &Repo{db: client, collection: collection}
}
