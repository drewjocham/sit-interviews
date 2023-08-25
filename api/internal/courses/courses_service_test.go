package courses

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
	Courses []Course
	Course  *Course
	err     error
}

func (m *MockRepository) CreateCourse(ctx context.Context, c *Course) error {
	return m.err
}

func (m *MockRepository) UpdateCourse(ctx context.Context, c *Course) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockRepository) DeleteCourse(ctx context.Context, id string) error {
	return m.err
}

func (m *MockRepository) GetAll(ctx context.Context, pipeline []bson.M) ([]Course, error) {
	return m.Courses, m.err
}

func (m *MockRepository) GetCourseById(ctx context.Context, id string) (*Course, error) {
	return m.Course, m.err
}

func Test_CoursesAllHandler(t *testing.T) {

	testCases := []struct {
		name        string
		repo        *MockRepository
		status      int
		resultCount int
		reqBody     []byte
		err         error
	}{
		{
			name: "Successfully get all courses",
			repo: &MockRepository{
				Courses: []Course{
					{
						Id: "123",
					},
					{
						Id: "456",
					},
				},
			},
			status: http.StatusOK,
			reqBody: []byte(`{
					"params": {
						"page": 1,
						"pageSize": 5
					},
					"data": {
						"all": true
					}
				}`),
			resultCount: 2,
			err:         nil,
		},
		{
			name: "Successfully test pagination",
			repo: &MockRepository{
				Courses: []Course{
					{
						Id: "123",
					},
					{
						Id: "456",
					},
					{
						Id: "457",
					},
					{
						Id: "452",
					},
				},
			},
			status: http.StatusOK,
			reqBody: []byte(`{
					"params": {
						"page": 1,
						"pageSize": 3
					},
					"data": {
						"all": true
					}
				}`),
			resultCount: 2,
			err:         nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			coursesService := NewCoursesService(tc.repo)

			req, err := http.NewRequest("POST", "/courses", bytes.NewBuffer(tc.reqBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(coursesService.CoursesAllHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.status, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.resultCount, len(response))

			// Assert other conditions as needed
			tc.repo.AssertExpectations(t)
		})
	}

}

// TODO: not sure why this 1 test is failing!!!!
func Test_CoursesIdHandler(t *testing.T) {

	testCases := []struct {
		name        string
		repo        *MockRepository
		status      int
		resultCount int
		id          string
		resBody     []byte
		err         error
	}{
		{
			name:        "Successfully get course by id",
			repo:        &MockRepository{},
			status:      http.StatusOK,
			resultCount: 1,
			id:          "123",
			err:         nil,
		},
		/*
			{
				name: "Unable to get course",
				repo: &MockRepository{
					Course: nil,
					err:    ErrUnableToGetCourse,
				},
				status:      http.StatusInternalServerError,
				resultCount: 0,
				id:          "555",
				err:         ErrUnableToGetCourse,
			},
		*/
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			coursesService := NewCoursesService(tc.repo)

			url := fmt.Sprintf("/v1/course/%s", tc.id)
			req, err := http.NewRequest("GET", url, nil)
			assert.ErrorIs(t, err, tc.err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()

			router.HandleFunc("/v1/course/{id}", coursesService.CoursesIdHandler).Methods("GET")
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.status, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)
		})
	}

}

func Test_CourseRepo_DeleteCourse(t *testing.T) {
	testCases := []struct {
		Name   string
		repo   *MockRepository
		Status int
		id     string
		Err    error
	}{
		{
			Name: "Successfully delete course",
			repo: &MockRepository{
				err: nil,
			},
			id:     "123",
			Status: http.StatusOK,
			Err:    nil,
		},
		{
			Name: "Delete course repo error",
			repo: &MockRepository{
				err: ErrDeleteCourse,
			},
			id:     "123",
			Status: http.StatusInternalServerError,
			Err:    ErrDeleteCourse,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {

			coursesService := NewCoursesService(tc.repo)

			url := fmt.Sprintf("/v1/delete-course/%s", tc.id)
			req, err := http.NewRequest("DELETE", url, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/v1/delete-course/{id}", coursesService.DeleteCourseHandler).Methods("DELETE")
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.Status, rr.Code)

		})
	}
}

func Test_CreateCourseHandler(t *testing.T) {
	testCases := []struct {
		name    string
		repo    *MockRepository
		status  int
		reqBody Course
		err     error
	}{
		{
			name:   "Successfully get create course",
			repo:   &MockRepository{},
			status: http.StatusOK,
			reqBody: Course{
				Id:           "course-id",
				Title:        "Sample Course",
				Description:  "This is a sample course",
				Lessons:      "10",
				Duration:     "2 weeks",
				Topic:        "Programming",
				Subscription: true,
				Details: struct {
					Title             string   `bson:"title" json:"title"`
					Instructor        string   `bson:"instructor" json:"instructor"`
					Introduction      string   `bson:"introduction" json:"introduction"`
					Learn             string   `bson:"learn" json:"learn"`
					Topics            []string `bson:"topics" json:"topics"`
					Prerequisites     string   `bson:"prerequisites" json:"prerequisites"`
					Goal              string   `bson:"goal" json:"goal"`
					AdditionalDetails string   `bson:"additionalDetails" json:"additionalDetails"`
					HighLevelOverview string   `bson:"highLevelOverview" json:"highLevelOverview"`
				}{
					Title:             "Sample Course",
					Instructor:        "John Doe",
					Introduction:      "Welcome to the sample course",
					Learn:             "Learn programming concepts",
					Topics:            []string{"Introduction", "Basics", "Advanced"},
					Prerequisites:     "None",
					Goal:              "Become a proficient programmer",
					AdditionalDetails: "Additional details about the course",
					HighLevelOverview: "Overview of the course",
				},
				Contents: []struct {
					SectionTitle string `bson:"section_title" json:"sectionTitle"`
					Videos       []struct {
						Title  string `bson:"title" json:"title"`
						Url    string `bson:"url" json:"url"`
						Paid   bool   `bson:"paid" json:"paid"`
						Length string `bson:"length" json:"length"`
					} `bson:"videos" json:"videos"`
				}{
					{
						SectionTitle: "Introduction",
						Videos: []struct {
							Title  string `bson:"title" json:"title"`
							Url    string `bson:"url" json:"url"`
							Paid   bool   `bson:"paid" json:"paid"`
							Length string `bson:"length" json:"length"`
						}{
							{
								Title:  "Welcome Video",
								Url:    "https://example.com/intro-video",
								Paid:   false,
								Length: "10 minutes",
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			name:   "Missing Topic, error",
			repo:   &MockRepository{},
			status: http.StatusBadRequest,
			reqBody: Course{
				Id:           "course-id",
				Title:        "Sample Course",
				Description:  "This is a sample course",
				Lessons:      "10",
				Duration:     "2 weeks",
				Topic:        "",
				Subscription: true,
				Details: struct {
					Title             string   `bson:"title" json:"title"`
					Instructor        string   `bson:"instructor" json:"instructor"`
					Introduction      string   `bson:"introduction" json:"introduction"`
					Learn             string   `bson:"learn" json:"learn"`
					Topics            []string `bson:"topics" json:"topics"`
					Prerequisites     string   `bson:"prerequisites" json:"prerequisites"`
					Goal              string   `bson:"goal" json:"goal"`
					AdditionalDetails string   `bson:"additionalDetails" json:"additionalDetails"`
					HighLevelOverview string   `bson:"highLevelOverview" json:"highLevelOverview"`
				}{
					Title:             "Sample Course",
					Instructor:        "John Doe",
					Introduction:      "Welcome to the sample course",
					Learn:             "Learn programming concepts",
					Topics:            []string{"Introduction", "Basics", "Advanced"},
					Prerequisites:     "None",
					Goal:              "Become a proficient programmer",
					AdditionalDetails: "Additional details about the course",
					HighLevelOverview: "Overview of the course",
				},
				Contents: []struct {
					SectionTitle string `bson:"section_title" json:"sectionTitle"`
					Videos       []struct {
						Title  string `bson:"title" json:"title"`
						Url    string `bson:"url" json:"url"`
						Paid   bool   `bson:"paid" json:"paid"`
						Length string `bson:"length" json:"length"`
					} `bson:"videos" json:"videos"`
				}{
					{
						SectionTitle: "Introduction",
						Videos: []struct {
							Title  string `bson:"title" json:"title"`
							Url    string `bson:"url" json:"url"`
							Paid   bool   `bson:"paid" json:"paid"`
							Length string `bson:"length" json:"length"`
						}{
							{
								Title:  "Welcome Video",
								Url:    "https://example.com/intro-video",
								Paid:   false,
								Length: "10 minutes",
							},
						},
					},
				},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			coursesService := NewCoursesService(tc.repo)
			mockCourseBytes, err := json.Marshal(tc.reqBody)
			assert.NoError(t, err)
			mockCourseJSON := []byte(fmt.Sprintf("%s", mockCourseBytes))

			req, err := http.NewRequest("POST", "/v1/create-course", bytes.NewReader(mockCourseJSON))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/v1/create-course", coursesService.CreateCourseHandler).Methods("POST")
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.status, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)
		})
	}
}
