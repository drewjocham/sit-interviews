package login

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"interviews/internal/auth"
	data "interviews/internal/users"
	"testing"
)

type MockRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
	user       data.User
}

func (m *MockRepository) Login(email string, password string) (data.User, error) {
	return data.User{}, nil
}

func TestLoginRepo_RegisterGoogle(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	testCases := []struct {
		name           string
		email          string
		claims         auth.GoogleClaims
		password       string
		mockRepo       *MockRepository
		mockedResponse bson.D
		err            error
	}{
		{
			name:  "user already registered with email",
			email: "test@test.com",
			mockRepo: &MockRepository{
				user: data.User{
					ID: int64(34),
				},
			},
			password: "test_test123",
			mockedResponse: bson.D{
				{"email", "test@test.com"},
			},
			claims: auth.GoogleClaims{
				Email:         "test@test.com",
				Name:          "test user",
				EmailVerified: false,
				FirstName:     "Test",
				LastName:      "Test1",
				StandardClaims: jwt.StandardClaims{
					Audience:  "",
					ExpiresAt: 0,
					Id:        "",
					IssuedAt:  0,
					Issuer:    "",
					NotBefore: 0,
					Subject:   "",
				},
			},
			err: ErrUserExists,
		},
		{
			name:  "New User Registers with Google Successfully",
			email: "test@test.com",
			mockRepo: &MockRepository{
				user: data.User{
					ID: int64(34),
				},
			},
			password:       "test_test123",
			mockedResponse: nil, // user not found
			claims: auth.GoogleClaims{
				Email:         "test@test.com",
				Name:          "test user",
				EmailVerified: false,
				FirstName:     "Test",
				LastName:      "Test1",
				StandardClaims: jwt.StandardClaims{
					Audience:  "",
					ExpiresAt: 0,
					Id:        "",
					IssuedAt:  0,
					Issuer:    "",
					NotBefore: 0,
					Subject:   "",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			mt.Run(tc.name, func(mt *mtest.T) {
				// findOne mock
				mt.AddMockResponses(mtest.CreateCursorResponse(0, "mock.users", mtest.FirstBatch, tc.mockedResponse))
				mt.AddMockResponses(mtest.CreateSuccessResponse())
				loginRepo := NewLoginRepository(mt.Client, mt.Coll)
				_, err := loginRepo.RegisterGoogle(tc.claims, tc.password)
				assert.ErrorIs(t, err, tc.err)
			})
		})

	}
}

func TestLoginRepo_Login(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	testCases := []struct {
		Name      string
		Email     string
		PassPlain string
		PassHash  string
		err       error
	}{
		{
			Name:      "Login successful",
			Email:     "test@test.com",
			PassPlain: "test_test123",
			PassHash:  "$2a$10$3LAcNwf/7rXdU1K2gRGKU.nOAeqM6B.CSKE7Nejd3VKCgoYZR/tYa",
			err:       nil,
		},
		{
			Name:      "invaild password hash",
			Email:     "test@test.com",
			PassPlain: "test_test123",
			PassHash:  "$2a$10$3LAcNwf/7rXdU1K2gRGKU.nM7th6R4.CSKE7Nejd3VKCgoYZR/tYa",
			err:       ErrInvalidPassword,
		},
	}

	for _, tc := range testCases {
		tc := tc
		// mt is needed to run the mongo mock
		mt.Run(tc.Name, func(mt *mtest.T) {
			// findOne mock
			mt.AddMockResponses(mtest.CreateCursorResponse(1, "mock.users", mtest.FirstBatch, bson.D{
				{"email", tc.Email},
				{"passHash", tc.PassHash},
			}))

			loginRepo := NewLoginRepository(mt.Client, mt.Coll)
			_, err := loginRepo.Login(tc.Email, tc.PassPlain)
			assert.ErrorIs(mt, err, tc.err)
		})
	}
}

func TestLoginRepo_Register(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	testCases := []struct {
		name           string
		email          string
		password       string
		mockedResponse bson.D
		err            error
	}{
		{
			name:     "Successful Registration",
			email:    "",
			password: "test_test123",
			mockedResponse: bson.D{
				{"ok", 1},
				{"email", ""},
			},
			err: nil,
		},
		{
			name:     "Unsuccessful (User Already Exists)",
			email:    "test@test2.com",
			password: "test_test123",
			mockedResponse: bson.D{
				{"email", "test@test.com"},
			},
			err: ErrUserExists,
		},
	}
	for _, tc := range testCases {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			mt.Run(tc.name, func(mt *mtest.T) {
				first := mtest.CreateCursorResponse(1, "mock.users",
					mtest.FirstBatch, tc.mockedResponse)
				second := mtest.CreateSuccessResponse()
				killCursors := mtest.CreateCursorResponse(0, "mock.users", mtest.NextBatch)
				mt.AddMockResponses(first, second, killCursors)

				loginRepo := NewLoginRepository(mt.Client, mt.Coll)
				_, err := loginRepo.Register(tc.email, tc.password)
				assert.ErrorIs(t, err, tc.err)
			})
		})
	}
}
