package login

import (
	"context"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"interviews/internal/auth"
	"interviews/internal/config"
	data "interviews/internal/users"
	"interviews/pkg"
	log "interviews/pkg/logger"
	validator "interviews/pkg/vaildator"
	"net/http"
	"strings"
)

type Login struct {
	helper    pkg.Helper
	user      data.User
	res       Response
	repo      Repository
	userRepo  UserRepository
	tokenRepo TokenRepository
	cfg       config.Config
	validator *validator.Validator
}

var (
	FailedLoginResponse = Response{
		Success: false,
		Message: "Login failed",
		Status:  http.StatusBadRequest,
		Token:   "",
	}
)

type Response struct {
	Success bool   `json:"success"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	User    string `json:"user"`
	Token   string `json:"token"`
	Error   string `json:"error"`
}

type Repository interface {
	Login(email string, password string) (data.User, error)
	Register(email, password string) (*data.User, error)
	RegisterGoogle(claims auth.GoogleClaims, password string) (*data.User, error)
}

type UserRepository interface {
	GetByEmail(email string) (*data.User, error)
}

type TokenRepository interface {
	GenerateJWT(ctx context.Context, user *data.User) (string, error)
	ValidateBearerToken(bearerToken string) (*jwt.Token, error)
	SaveToken(ctx context.Context, token string, email string) error
	ValidateGoogleJWT(tokenString string) (auth.GoogleClaims, error)
	DeleteToken(ctx context.Context, email string) error
	GetEmailFromJWT(tokenString string) (string, error)
}

type TokenCache interface {
	AddToken(ctx context.Context, email, token string) error
	RemoveToken(ctx context.Context, email string) error
}

func (l *Login) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	decoder := json.NewDecoder(r.Body)
	var requestBody map[string]string
	err := decoder.Decode(&requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	email, ok := requestBody["email"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	password, ok := requestBody["password"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	l.ValidateUser(l.validator, email, password)
	l.validateEmail(l.validator, email)
	//TODO: have a look at why the password length fails sometimes when it is still valid
	l.validatePasswordPlaintext(l.validator, password)

	if l.validator.Valid() {
		user, err := l.repo.Login(email, password)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(FailedLoginResponse)

			return
		}

		token, err := l.tokenRepo.GenerateJWT(ctx, &user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(FailedLoginResponse)

			return
		}

		response := Response{
			Success: true,
			Message: "Login successful",
			User:    email,
			Token:   token,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FailedLoginResponse)
	}

}
func (l *Login) LoginGoogleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	decoder := json.NewDecoder(r.Body)
	var requestBody map[string]string
	err := decoder.Decode(&requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	token, ok := requestBody["token"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	claims, err := l.tokenRepo.ValidateGoogleJWT(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	user, err := l.userRepo.GetByEmail(claims.Email)

	l.validateEmail(l.validator, claims.Email)

	if err != nil && l.validator.Valid() {

		var err error
		// no password needed if login with Google
		user, err = l.repo.RegisterGoogle(claims, "password")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(FailedLoginResponse)

			return
		}

	}

	validToken, err := l.tokenRepo.GenerateJWT(ctx, user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	response := Response{
		Success: true,
		Message: "Login successful",
		Token:   validToken,
	}

	json.NewEncoder(w).Encode(response)

}
func (l *Login) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	decoder := json.NewDecoder(r.Body)
	var requestBody map[string]string
	err := decoder.Decode(&requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	email, ok := requestBody["email"]
	if !ok || email == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	password, ok := requestBody["password"]
	if !ok || password == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	//TODO: add validation for email and password

	user, err := l.repo.Register(email, password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	token, err := l.tokenRepo.GenerateJWT(ctx, user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	response := map[string]interface{}{
		"message": "Account created successfully",
		"token":   token,
	}

	json.NewEncoder(w).Encode(response)
}
func (l *Login) PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
func (l *Login) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
func (l *Login) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	clog := log.GetLoggerFromContext(ctx)

	w.Header().Add("Vary", "Authorization")

	authorizationHeader := r.Header.Get("Authorization")

	if authorizationHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	// TODO: put token in a blacklist and clear the list every 24 hours
	// or use token versioning
	bearerToken := headerParts[1]
	clog.InfoCtx(bearerToken, log.Ctx{
		"msg": "Put this is a blacklist",
	})

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		json.NewEncoder(w).Encode(FailedLoginResponse)

		return
	}

	// If there is no logout-related operation to perform, you can simply respond with success.
	successResponse := Response{
		Success: true,
		Message: "Logout successful",
		Token:   "",
	}

	json.NewEncoder(w).Encode(successResponse)
}
func (l *Login) ValidateUser(v *validator.Validator, email, password string) {
	l.validateEmail(v, email)
	l.validatePasswordPlaintext(v, password)
}
func (l *Login) validateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}
func (l *Login) validatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	//v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
func NewLoginService(repository Repository,
	tokenRepo TokenRepository,
	userRepo UserRepository) *Login {
	return &Login{
		helper:    pkg.Helper{},
		repo:      repository,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		validator: validator.New(),
	}
}
