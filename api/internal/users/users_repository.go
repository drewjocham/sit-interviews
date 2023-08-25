package data

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	validator "interviews/pkg/vaildator"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrInvalidEmail   = errors.New("invalid email")
)

var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id" bson:"id"`
	CreatedAt time.Time `json:"created_at" bson:"createdAt"`
	Name      string    `json:"name" bson:"name"`
	Email     string    `json:"email" bson:"email"`
	PassPlain string    `json:"passPlain" bson:"passPlain"`
	PassHash  string    `json:"passHash" bson:"passHash"`
	Role      string    `json:"role" bson:"role"`
	Activated bool      `json:"activated" bson:"activated"`
	Version   int       `json:"-" bson:"version"`
}

type UserRepo struct {
	db         *mongo.Client
	collection *mongo.Collection
	tokenRepo  TokenRepository
	valid      *validator.Validator
}

type TokenRepository interface {
	ValidateBearerToken(bearerToken string) (*jwt.Token, error)
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type Password struct {
	Plaintext *string
	Hash      []byte
}

func (u *UserRepo) GetByEmail(email string) (*User, error) {
	ctx := context.Background()

	u.ValidateEmail(u.valid, email)
	if !u.valid.Valid() {
		return &User{}, ErrInvalidEmail
	}

	var user *User
	err := u.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return &User{}, err
	}

	return user, nil
}

func (u *UserRepo) Update(ctx context.Context, user *User) error {

	update := bson.M{
		"$set": bson.M{
			"name":      user.Name,
			"email":     user.Email,
			"passPlain": user.PassPlain,
			"passHash":  user.PassHash,
			"role":      user.Role,
			"activated": user.Activated,
			"version":   user.Version,
		},
	}

	filter := bson.M{"email": user.Email}
	_, err := u.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepo) UpdateRole(ctx context.Context, user *User) error {
	update := bson.M{
		"$set": bson.M{
			"role": user.Role,
		},
	}

	filter := bson.M{"email": user.Email}
	_, err := u.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	return nil
}

func (p *Password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.Plaintext = &plaintextPassword
	p.Hash = hash

	return nil
}

func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (u *UserRepo) ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func (u *UserRepo) ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func (u *UserRepo) ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	u.ValidateEmail(v, user.Email)

	if user.PassPlain != "" {
		u.ValidatePasswordPlaintext(v, user.PassPlain)
	}

}

func NewUserRepository(client *mongo.Client, collection *mongo.Collection,
	tokenRepo TokenRepository) *UserRepo {
	return &UserRepo{
		db:         client,
		collection: collection,
		tokenRepo:  tokenRepo,
		valid:      validator.New(),
	}
}
