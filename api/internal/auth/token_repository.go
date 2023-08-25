package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	data "interviews/internal/users"
	log "interviews/pkg/logger"
	validator "interviews/pkg/vaildator"
	"io"
	"net/http"
	"time"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
	ScopePasswordReset  = "password-reset"
	TokenSecret         = "NLp948i7pvQbgTYrGHtpVFaSUrkbxqpqnqf3hdR-ves="
	// Not a valid GoogleClientID
	GoogleClientID = "705206800363-b7kutdfkhfd8r8ge76it0t7ur6p4pmr25p.apps.googleusercontent.com"
)

var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("invalid token")
)

type GoogleClaims struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	jwt.StandardClaims
}

type TokenRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type UserRepository interface {
	GetByEmail(email string) (*data.User, error)
}

type Token struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
	Email     string    `bson:"email" json:"email"`
	Token     string    `bson:"token" json:"token"`
	Role      string    `bson:"role" json:"role"`
}

func (m *TokenRepository) GetAllTokens(ctx context.Context) (map[string]string, error) {
	clog := log.GetLoggerFromContext(ctx)

	cur, err := m.collection.Find(ctx, bson.D{{}})
	if err != nil {
		clog.Error(err)

		return nil, err
	}
	defer cur.Close(ctx)

	var tokens []Token

	err = cur.All(ctx, &tokens)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred while retrieving for all courses"})

		return nil, err
	}

	// Create a map of tokens
	tm := make(map[string]string)
	for _, t := range tokens {
		tm[t.Email] = t.Token
	}

	return tm, nil
}

func (m *TokenRepository) SaveToken(ctx context.Context, token, email string) error {

	// Insert or update the token
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"token": token, "email": email}}

	_, err := m.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {

		return err
	}

	return nil
}

func (m *TokenRepository) GetToken(ctx context.Context, email string) (string, error) {
	filter := bson.M{"email": email}

	var t Token

	err := m.collection.FindOne(ctx, filter).Decode(&t)
	if err != nil {
		return "", err
	}

	return t.Token, nil
}

func (m *TokenRepository) TokenExists(ctx context.Context, email, token string) (bool, error) {
	filter := bson.M{"email": email}

	var savedToken Token

	err := m.collection.FindOne(ctx, filter).Decode(&savedToken)
	if err != nil {
		return false, err
	}

	return token == savedToken.Token, nil
}

func (m *TokenRepository) DeleteToken(ctx context.Context, email string) error {
	filter := bson.M{"email": email}

	_, err := m.collection.DeleteMany(ctx, filter)

	if err != nil {
		return err
	}

	return nil
}

func (m *TokenRepository) ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) bool {
	v.Check(tokenPlaintext != "", "token", "must be provided")

	_, err := m.ValidateBearerToken(tokenPlaintext)
	if err != nil {
		v.AddError("token", "invalid token")

		return false
	}

	return true
}

func (m *TokenRepository) GenerateJWT(ctx context.Context, user *data.User) (string, error) {
	clog := log.GetLoggerFromContext(ctx)

	expirationTime := time.Now().Add(time.Hour * 24)
	payload := make(map[string]interface{})

	claims := jwt.MapClaims{}

	claims["jti"] = uuid.NewString()
	claims["exp"] = expirationTime.Unix()
	claims["email"] = user.Email
	claims["role"] = user.Role
	claims["aud"] = "go-trakteer"

	for i, v := range payload {
		claims[i] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(TokenSecret))
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred while generating jwt token"})
	}

	return tokenString, nil
}

func (m *TokenRepository) ValidateBearerToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return []byte(TokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	ok := m.isExpired(token)
	if !ok {
		return nil, ErrExpiredToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}

func (m *TokenRepository) isExpired(token *jwt.Token) bool {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return true // Invalid claims format
	}
	expiryTimeFloat, ok := claims["exp"].(float64)
	if !ok {
		return true // Invalid expiry time format
	}
	expiryTime := int64(expiryTimeFloat)
	return time.Now().Unix() > expiryTime
}

func (m *TokenRepository) DecodeToken(accessToken *jwt.Token) Token {
	var token Token
	stringify, _ := json.Marshal(&accessToken)
	json.Unmarshal(stringify, &token)

	return token
}

func (m *TokenRepository) GetEmailFromJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(TokenSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email, emailOk := claims["email"].(string)
		if !emailOk {
			return "", fmt.Errorf("invalid token claims")
		}
		return email, nil
	}

	return "", fmt.Errorf("invalid token")
}

func getGooglePublicKey(keyID string) (string, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		return "", err
	}

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	myResp := map[string]string{}
	err = json.Unmarshal(dat, &myResp)
	if err != nil {
		return "", err
	}

	key, ok := myResp[keyID]
	if !ok {
		return "", errors.New("key not found")
	}

	return key, nil
}

func (m *TokenRepository) ValidateGoogleJWT(tokenString string) (GoogleClaims, error) {
	claimsStruct := GoogleClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
			pem, err := getGooglePublicKey(fmt.Sprintf("%s", token.Header["kid"]))
			if err != nil {
				return nil, err
			}
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
			if err != nil {
				return nil, err
			}
			return key, nil
		},
	)
	if err != nil {
		return GoogleClaims{}, err
	}

	claims, ok := token.Claims.(*GoogleClaims)
	if !ok {
		return GoogleClaims{}, errors.New("Invalid Google JWT")
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return GoogleClaims{}, errors.New("iss is invalid")
	}

	if claims.Audience != GoogleClientID {
		return GoogleClaims{}, errors.New("aud is invalid")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return GoogleClaims{}, errors.New("JWT is expired")
	}

	return *claims, nil
}

func NewTokenRepository(client *mongo.Client,
	collection *mongo.Collection) *TokenRepository {
	return &TokenRepository{
		client:     client,
		collection: collection,
	}
}
