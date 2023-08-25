package login

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"interviews/internal/auth"
	data "interviews/internal/users"
	validator "interviews/pkg/vaildator"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	VALID_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJnby10cmFrdGVlciIsImVtYWlsIjoiY29udGFjdEBqb2NoYW0uaW8iLCJleHAiOjE2OTE5OTQ3MzAsImp0aSI6Ijc2ODEyYTEzLWMyN2UtNGRlOS1hMDA5LTE3M2EzY2RkODI2NCIsInJvbGUiOiIifQ.ZNoDLSTvqZm6bqxbHWmyHl6hlI-xJf0AvnNwqrZA4qQ"
)

type MockLoginRepository struct {
	mock.Mock
	user  *data.User
	error error
}

func (m *MockLoginRepository) Register(email, password string) (*data.User, error) {
	return m.user, m.error
}

func (m *MockLoginRepository) RegisterGoogle(claims auth.GoogleClaims, password string) (*data.User, error) {

	// Simulate password hashing
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create a new user in the mock database
	newUser := &data.User{
		Name:      claims.Name,
		Email:     claims.Email,
		PassHash:  hash,
		CreatedAt: time.Now(),
		Role:      "",
		Activated: false,
	}

	return newUser, nil
}

func (m *MockLoginRepository) Login(email string, password string) (data.User, error) {
	return *m.user, m.error
}

type MockTokenRepository struct {
	token string
}

type MockUserRepository struct {
	user *data.User
}

func (m *MockUserRepository) GetByEmail(email string) (*data.User, error) {
	return m.user, nil
}

func (m *MockTokenRepository) ValidateBearerToken(bearerToken string) (*jwt.Token, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTokenRepository) SaveToken(ctx context.Context, token string, email string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTokenRepository) ValidateGoogleJWT(tokenString string) (auth.GoogleClaims, error) {
	sampleClaims := auth.GoogleClaims{
		Name:          "John Doe",
		Email:         "johndoe@example.com",
		EmailVerified: true,
	}

	if tokenString == "valid_mock_token" {
		return sampleClaims, nil
	}

	// Simulate an invalid token scenario for demonstration purposes
	return auth.GoogleClaims{}, errors.New("invalid token")
}

func (m *MockTokenRepository) DeleteToken(ctx context.Context, email string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTokenRepository) GetEmailFromJWT(tokenString string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTokenRepository) GenerateJWT(ctx context.Context, user *data.User) (string, error) {
	return m.token, nil
}
func Test_LoginHandler(t *testing.T) {

	testCases := []struct {
		Name             string
		login            *Login
		requestBody      map[string]string
		expectedResponse Response
		statusCode       int
		token            string
		user             data.User
	}{
		{
			Name: "Missing Password",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{
						ID:        0,
						CreatedAt: time.Time{},
						Name:      "Test",
						Email:     "",
						PassPlain: "asdfasdfadsf",
						PassHash:  "",
						Role:      "",
						Activated: false,
						Version:   0,
					},
				},
				tokenRepo: &MockTokenRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "",
			},
			statusCode:       http.StatusBadRequest,
			expectedResponse: FailedLoginResponse,
		},
		{
			Name: "Missing Email",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{
						ID:        0,
						CreatedAt: time.Time{},
						Name:      "Test",
						Email:     "",
						PassPlain: "yxcy<ycx<",
						PassHash:  "",
						Role:      "",
						Activated: false,
						Version:   0,
					},
				},
				tokenRepo: &MockTokenRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"email":    "",
				"password": "testpassword",
			},
			statusCode:       http.StatusBadRequest,
			expectedResponse: FailedLoginResponse,
		},
		{
			Name: "Invalid Request Body",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{
						ID:        0,
						CreatedAt: time.Time{},
						Name:      "Test",
						Email:     "test@test.com",
						PassPlain: "asdfasdfadsf",
						PassHash:  "",
						Role:      "",
						Activated: false,
						Version:   0,
					},
				},
				tokenRepo: &MockTokenRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"email":    "",
				"password": "testpassword",
			},
			expectedResponse: FailedLoginResponse,
			statusCode:       http.StatusBadRequest,
		},
		{
			Name: "Successful Login",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{
						ID:        0,
						CreatedAt: time.Time{},
						Name:      "Test",
						Email:     "test@test.com",
						PassPlain: "asdfasdfadsf",
						PassHash:  "",
						Role:      "",
						Activated: false,
						Version:   0,
					},
				},
				tokenRepo: &MockTokenRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"email":    "test@test.com",
				"password": "testpassword",
			},
			expectedResponse: Response{
				Success: true,
				Message: "Login successful",
				User:    "test@test.com",
				Token:   "testtoken",
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			requestBodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBodyBytes))
			tc.login.LoginHandler(recorder, req)

			// Check the status code
			assert.Equal(t, tc.statusCode, recorder.Result().StatusCode)

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResponse.Message, response.Message)
			assert.Equal(t, tc.expectedResponse.User, response.User)
		})
	}

}

func TestLogin_LoginGoogleHandler(t *testing.T) {

	testCases := []struct {
		Name             string
		login            *Login
		requestBody      map[string]string
		expectedResponse Response
		statusCode       int
		token            string
		user             data.User
	}{
		{
			Name: "Missing Token",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{
						ID:        0,
						CreatedAt: time.Time{},
						Name:      "Test",
						Email:     "",
						PassPlain: "asdfasdfadsf",
						PassHash:  "",
						Role:      "",
						Activated: false,
						Version:   0,
					},
				},
				tokenRepo: &MockTokenRepository{},
				userRepo:  &MockUserRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"token": "",
			},
			statusCode:       http.StatusBadRequest,
			expectedResponse: FailedLoginResponse,
		},
		{
			Name: "Successful login",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{
						ID:        0,
						CreatedAt: time.Time{},
						Name:      "Test",
						Email:     "",
						PassPlain: "",
						PassHash:  "",
						Role:      "",
						Activated: false,
						Version:   0,
					},
				},
				tokenRepo: &MockTokenRepository{},
				userRepo:  &MockUserRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"token": "valid_mock_token",
			},
			statusCode: http.StatusOK,
			expectedResponse: Response{
				Success: true,
				Message: "Login successful",
				Token:   VALID_TOKEN,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			requestBodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/v1/google", bytes.NewReader(requestBodyBytes))
			tc.login.LoginGoogleHandler(recorder, req)

			// Check the status code
			assert.Equal(t, tc.statusCode, recorder.Result().StatusCode)

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResponse.Message, response.Message)
			assert.Equal(t, tc.expectedResponse.User, response.User)

		})
	}
}

func TestLogin_CreateAccountHandler(t *testing.T) {

	testCases := []struct {
		Name             string
		login            *Login
		requestBody      map[string]string
		expectedResponse Response
		statusCode       int
		user             data.User
	}{
		{
			Name: "Missing Email",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{},
				},
				tokenRepo: &MockTokenRepository{},
				userRepo:  &MockUserRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"email":    "",
				"password": "testpassword",
			},
			statusCode:       http.StatusBadRequest,
			expectedResponse: FailedLoginResponse,
		},
		{
			Name: "Missing Password",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{},
				},
				tokenRepo: &MockTokenRepository{},
				userRepo:  &MockUserRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"email":    "test@test.com",
				"password": "",
			},
			statusCode:       http.StatusBadRequest,
			expectedResponse: FailedLoginResponse,
		},
		{
			Name: "Successful Account Creation",
			login: &Login{
				repo: &MockLoginRepository{
					user: &data.User{},
				},
				tokenRepo: &MockTokenRepository{},
				userRepo:  &MockUserRepository{},
				validator: validator.New(),
			},
			requestBody: map[string]string{
				"email":    "test@test.com",
				"password": "Test/67$PassWord",
			},
			statusCode: http.StatusOK,
			expectedResponse: Response{
				Success: true,
				Message: "Account created successfully",
				Token:   VALID_TOKEN,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			requestBodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/v1/create-account", bytes.NewReader(requestBodyBytes))
			tc.login.CreateAccountHandler(recorder, req)

			// Check the status code
			assert.Equal(t, tc.statusCode, recorder.Result().StatusCode)

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResponse.Message, response.Message)
			assert.Equal(t, tc.expectedResponse.User, response.User)
		})
	}
}
