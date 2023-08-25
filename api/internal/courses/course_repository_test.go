package courses

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCourseRepo_GetAll(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	testCases := []struct {
		name           string
		mockedResponse []bson.D
		expected       int
		err            error
	}{
		{
			name:     "GetAll Courses Successfully",
			expected: 2,
			mockedResponse: []bson.D{
				{
					{"ok", 1},
					{"value", bson.D{
						{"_id", primitive.NewObjectID()},
						{"title", "Sample Course"},
						{"description", "This is a sample course"},
						{"lessons", "10"},
						{"duration", "2 weeks"},
						{"topic", "Programming"},
						{"subscription", true},
						{"course_details", bson.D{
							{"title", "Sample Course"},
							{"instructor", "John Doe"},
							{"introduction", "Welcome to the sample course"},
							{"learn", "Learn programming concepts"},
							{"topics", bson.A{"Introduction", "Basics", "Advanced"}},
							{"prerequisites", "None"},
							{"goal", "Become a proficient programmer"},
							{"additionalDetails", "Additional details about the course"},
							{"highLevelOverview", "Overview of the course"},
						}},
						{"contents", bson.A{
							bson.D{
								{"section_title", "Introduction"},
								{"videos", bson.A{
									bson.D{
										{"title", "Welcome Video"},
										{"url", "https://example.com/intro-video"},
										{"paid", false},
										{"length", "10 minutes"},
									},
								}},
							},
						}},
					}},
				},
				{
					{"ok", 1},
					{"value", bson.D{
						{"_id", primitive.NewObjectID()},
						{"title", "Sample Course"},
						{"description", "This is a sample course"},
						{"lessons", "10"},
						{"duration", "8 weeks"},
						{"topic", "Programming"},
						{"subscription", true},
						{"course_details", bson.D{
							{"title", "Sample Course In Go"},
							{"instructor", "John Poo"},
							{"introduction", "Welcome to the sample course"},
							{"learn", "Learn programming concepts"},
							{"topics", bson.A{"Introduction", "Basics", "Advanced"}},
							{"prerequisites", "None"},
							{"goal", "Become a proficient programmer"},
							{"additionalDetails", "Additional details about the course"},
							{"highLevelOverview", "Overview of the course"},
						}},
						{"contents", bson.A{
							bson.D{
								{"section_title", "Introduction"},
								{"videos", bson.A{
									bson.D{
										{"title", "Welcome Video"},
										{"url", "https://example.com/intro-video"},
										{"paid", false},
										{"length", "10 minutes"},
									},
								}},
							},
						}},
					}},
				},
			},
			err: nil,
		},
		{
			name:     "Error Occurred while Getting All Courses",
			expected: 0,
			mockedResponse: []bson.D{
				{
					{"ok", 0},
				},
			},
			err: ErrGetAllCourses,
		},
	}
	for _, tc := range testCases {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			if tc.err == nil {
				first := mtest.CreateCursorResponse(1, "mock.courses", mtest.FirstBatch, tc.mockedResponse[0])
				getMore := mtest.CreateCursorResponse(1, "mock.courses", mtest.NextBatch, tc.mockedResponse[1])
				killCursors := mtest.CreateCursorResponse(0, "mock.courses", mtest.NextBatch)
				mt.AddMockResponses(first, getMore, killCursors)
			} else {
				first := mtest.CreateCursorResponse(1, "mock.courses", mtest.FirstBatch, tc.mockedResponse[0])
				mt.AddMockResponses(first, mtest.CreateWriteErrorsResponse())
			}

			repo := NewCourseRepository(mt.Client, mt.Coll)

			ctx := context.Background()
			pipeline := []bson.M{}

			courses, err := repo.GetAll(ctx, pipeline)

			assert.Equal(t, tc.expected, len(courses))
			assert.ErrorIs(t, err, tc.err)
		})
	}

}

func TestCourseRepo_GetCourseById(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	testCases := []struct {
		name           string
		mockedResponse bson.D
		expected       int
		courseID       string
		err            error
	}{
		{
			name: "Successfully GetOne Course By ID",
			mockedResponse: bson.D{
				{"ok", 1},
				{"value", bson.D{
					{"_id", primitive.NewObjectID()},
					{"title", "Sample Course"},
					{"description", "This is a sample course"},
					{"lessons", "10"},
					{"duration", "2 weeks"},
					{"topic", "Programming"},
					{"subscription", true},
					{"course_details", bson.D{
						{"title", "Sample Course"},
						{"instructor", "John Doe"},
						{"introduction", "Welcome to the sample course"},
						{"learn", "Learn programming concepts"},
						{"topics", bson.A{"Introduction", "Basics", "Advanced"}},
						{"prerequisites", "None"},
						{"goal", "Become a proficient programmer"},
						{"additionalDetails", "Additional details about the course"},
						{"highLevelOverview", "Overview of the course"},
					}},
					{"contents", bson.A{
						bson.D{
							{"section_title", "Introduction"},
							{"videos", bson.A{
								bson.D{
									{"title", "Welcome Video"},
									{"url", "https://example.com/intro-video"},
									{"paid", false},
									{"length", "10 minutes"},
								},
							}},
						},
					}},
				}},
			},
			courseID: primitive.NewObjectID().Hex(),
			expected: 1,
		},
		{
			name: "Error converting hex",
			mockedResponse: bson.D{
				{"ok", 1},
				{"value", bson.D{
					{"_id", primitive.NewObjectID()},
					{"title", "Sample Course"},
					{"description", "This is a sample course"},
					{"lessons", "10"},
					{"duration", "2 weeks"},
					{"topic", "Programming"},
					{"subscription", true},
					{"details", bson.D{
						{"title", "Sample Course"},
						{"instructor", "John Doe"},
						{"introduction", "Welcome to the sample course"},
						{"learn", "Learn programming concepts"},
						{"topics", bson.A{"Introduction", "Basics", "Advanced"}},
						{"prerequisites", "None"},
						{"goal", "Become a proficient programmer"},
						{"additionalDetails", "Additional details about the course"},
						{"highLevelOverview", "Overview of the course"},
					}},
					{"contents", bson.A{
						bson.D{
							{"section_title", "Introduction"},
							{"videos", bson.A{
								bson.D{
									{"title", "Welcome Video"},
									{"url", "https://example.com/intro-video"},
									{"paid", false},
									{"length", "10 minutes"},
								},
							}},
						},
					}},
				}},
			},
			courseID: "fail",
			expected: 0,
			err:      ErrConvertingKeyToHex,
		},
	}
	for _, tc := range testCases {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateCursorResponse(1, "mock.courses",
				mtest.FirstBatch, tc.mockedResponse))

			repo := NewCourseRepository(mt.Client, mt.Coll)

			ctx := context.Background()

			course, err := repo.GetCourseById(ctx, tc.courseID)
			assert.ErrorIs(t, err, tc.err)
			if err == nil {
				assert.NotNil(t, course)
			}
		})
	}
}

func TestCourseRepo_CreateCourse(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	testCases := []struct {
		name         string
		newCourse    *Course
		mockResponse bson.D
		err          error
	}{
		{
			name: "Successful Creation of Course ",
			newCourse: &Course{
				Title:       "New Course",
				Description: "Description of the new course",
			},
			mockResponse: bson.D{{"ok", 1}},
			err:          nil,
		},
		{
			name: "Error In Creation of Course ",
			newCourse: &Course{
				Title:       "New Course",
				Description: "Description of the new course",
			},
			mockResponse: bson.D{{"ok", 0}},
			err:          ErrCreateCourse,
		},
	}
	for _, tc := range testCases {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			// InsertOne requires only a success response.
			mt.AddMockResponses(tc.mockResponse)

			repo := NewCourseRepository(mt.Client, mt.Coll)

			ctx := context.Background()

			err := repo.CreateCourse(ctx, tc.newCourse)
			assert.ErrorIs(t, err, tc.err)
		})
	}

}

func TestCourseRepo_UpdateCourse(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	testCases := []struct {
		name         string
		course       *Course
		mockResponse bson.D
		expected     int
		err          error
	}{
		{
			name: "Update Course Successful",
			course: &Course{
				Id:          id,
				Title:       "I am a old title",
				Description: "test",
				Lessons:     "3",
				Duration:    "5 hours",
			},
			mockResponse: bson.D{
				{"ok", 1},
				{"value", bson.D{
					{"_id", id},
					{"title", "Test Successful"},
					{"description", "This is a sample course"},
					{"lessons", "10"},
					{"duration", "2 weeks"},
					{"topic", "Test"},
					{"subscription", true},
				},
				},
			},
			err: nil,
		},
		/*
			{
				name: "Update Course Error",
				course: &Course{
					Id:          primitive.NewObjectID().Hex(),
					Title:       "I am a old title",
					Description: "test",
					Lessons:     "3",
					Duration:    "5 hours",
				},
				mockResponse: bson.D{
					{"ok", 0},
				},
				err: ErrUpdateCourse,
			},
		*/
	}
	for _, tc := range testCases {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateCursorResponse(1, "mock.course", mtest.FirstBatch, tc.mockResponse))
			//mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})

			repo := NewCourseRepository(mt.Client, mt.Coll)

			err := repo.UpdateCourse(ctx, tc.course)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestCourseRepo_DeleteCourse(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	testCases := []struct {
		name             string
		courseIDToDelete string
		mockResponse     bson.D
		err              error
	}{
		{
			name:             "Successful Delete",
			courseIDToDelete: primitive.NewObjectID().Hex(),
			mockResponse: bson.D{
				{"ok", 1},
				{"acknowledged", true},
				{"n", 1},
			},
			err: nil,
		},
		{
			name:             "Error Delete",
			courseIDToDelete: primitive.NewObjectID().Hex(),
			mockResponse: bson.D{
				{"ok", 0},
				{"acknowledged", false},
				{"n", 1},
			},
			err: ErrDeleteCourse,
		},
	}
	for _, tc := range testCases {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			mt.AddMockResponses(tc.mockResponse)
			repo := NewCourseRepository(mt.Client, mt.Coll)

			ctx := context.Background()

			err := repo.DeleteCourse(ctx, tc.courseIDToDelete)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
