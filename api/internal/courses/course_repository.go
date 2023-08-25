package courses

import (
	"context"
	"errors"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres" // selecting goqu dialect
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	log "interviews/pkg/logger"
)

const (
	noPercentage = 100
	schema       = "interviewsapi"
)

var (
	ErrGetAllCourses      = errors.New("error occurred while retrieving for all courses")
	ErrConvertingKeyToHex = errors.New("error occurred in GetCourseById while converting object key to hex")
	ErrUpdateCourse       = errors.New("error occurred in UpdateCourse while updating record")
	ErrCourseNotFound     = errors.New("course not found")
)

type Course struct {
	ObjectId     primitive.ObjectID `bson:"_id,omitempty" json:"objectId"`
	Id           string             `bson:"id" json:"id"`
	Title        string             `bson:"title" json:"title"`
	Description  string             `bson:"description" json:"description"`
	Lessons      string             `bson:"lessons" json:"lessons"`
	Duration     string             `bson:"duration" json:"duration"`
	Topic        string             `bson:"topic" json:"topic"`
	Subscription bool               `bson:"subscription" json:"subscription"`
	Details      struct {
		Title             string   `bson:"title" json:"title"`
		Instructor        string   `bson:"instructor" json:"instructor"`
		Introduction      string   `bson:"introduction" json:"introduction"`
		Learn             string   `bson:"learn" json:"learn"`
		Topics            []string `bson:"topics" json:"topics"`
		Prerequisites     string   `bson:"prerequisites" json:"prerequisites"`
		Goal              string   `bson:"goal" json:"goal"`
		AdditionalDetails string   `bson:"additionalDetails" json:"additionalDetails"`
		HighLevelOverview string   `bson:"highLevelOverview" json:"highLevelOverview"`
	} `bson:"course_details" json:"course_details"`
	Contents []struct {
		SectionTitle string `bson:"section_title" json:"sectionTitle"`
		Videos       []struct {
			Title  string `bson:"title" json:"title"`
			Url    string `bson:"url" json:"url"`
			Paid   bool   `bson:"paid" json:"paid"`
			Length string `bson:"length" json:"length"`
		} `bson:"videos" json:"videos"`
	} `bson:"contents" json:"contents"`
}

type CourseRepo struct {
	db         *mongo.Client
	collection *mongo.Collection
}

func (r *CourseRepo) GetAll(ctx context.Context, pipeline []bson.M) ([]Course, error) {
	clog := log.GetLoggerFromContext(ctx)

	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)

		return nil, err
	}
	defer cur.Close(ctx)

	var courses []Course

	err = cur.All(ctx, &courses)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred while retrieving for all courses"})

		return nil, ErrGetAllCourses
	}

	return courses, nil
}

func (r *CourseRepo) GetCourseById(ctx context.Context, id string) (*Course, error) {
	clog := log.GetLoggerFromContext(ctx)

	var course *Course

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred in GetCourseById while converting object key to hex"})

		return course, ErrConvertingKeyToHex
	}

	filter := bson.M{"_id": objID}

	err = r.collection.FindOne(ctx, filter).Decode(&course)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "an error occurred while finding a course"})

		return course, err
	}

	return course, nil
}

func (r *CourseRepo) CreateCourse(ctx context.Context, c *Course) error {

	clog := log.GetLoggerFromContext(ctx)

	_, err := r.collection.InsertOne(ctx, c)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred while creating new course"})

		return ErrCreateCourse
	}

	return nil
}

// TODO: should return the change
func (r *CourseRepo) UpdateCourse(ctx context.Context, c *Course) error {
	clog := log.GetLoggerFromContext(ctx)

	objID, err := primitive.ObjectIDFromHex(c.Id)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred in UpdateCourse while converting object key to hex"})

		return ErrConvertingKeyToHex
	}

	filter := bson.D{
		{"_id", objID},
	}

	update := bson.D{
		{"$set", bson.D{
			{"title", c.Title},
			{"description", c.Description},
			{"lessons", c.Lessons},
			{"duration", c.Duration},
		}},
	}

	var updatedCourse Course
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedCourse)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrCourseNotFound
		}
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred in UpdateCourse while updating record"})
		return ErrUpdateCourse
	}

	return nil
}

func (r *CourseRepo) DeleteCourse(ctx context.Context, id string) error {
	clog := log.GetLoggerFromContext(ctx)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred in DeleteCourse while converting object key to hex"})

		return ErrConvertingKeyToHex
	}

	filter := bson.D{
		{"_id", objID},
	}

	_, err = r.collection.DeleteOne(ctx, filter)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{"msg": "error occurred in DeleteCourse while deleting record"})

		return ErrDeleteCourse
	}

	return nil
}

func NewCourseRepository(client *mongo.Client, collection *mongo.Collection) *CourseRepo {
	return &CourseRepo{
		db:         client,
		collection: collection,
	}
}
