package courses

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"interviews/pkg"
	log "interviews/pkg/logger"
	validator "interviews/pkg/vaildator"
	"net/http"
)

var (
	ErrDecodingCourse      = errors.New("error occurred while decoding the course object")
	ErrBadRequestBody      = errors.New("bad request body")
	ErrGetFilter           = errors.New("error occurred while getting filter")
	ErrMissingID           = errors.New("id is missing in parameters")
	ErrUnableToGetCourse   = errors.New("unable to get course")
	ErrDeleteCourse        = errors.New("an error occurred while delete a course")
	ErrCreateCourse        = errors.New("an error occurred while inserting course")
	ErrInvalidCourseFormat = errors.New("course is missing required fields")
)

type ErrResponse struct {
	Message error `json:"message"`
}

type Repository interface {
	GetAll(ctx context.Context, pipeline []bson.M) ([]Course, error)
	GetCourseById(ctx context.Context, id string) (*Course, error)
	CreateCourse(ctx context.Context, c *Course) error
	UpdateCourse(ctx context.Context, c *Course) error
	DeleteCourse(ctx context.Context, id string) error
}

type Courses struct {
	helper    pkg.Helper
	repo      Repository
	validator validator.Validator
}

type Filter struct {
	All          bool `json:"all"`
	Beginner     bool `json:"beginner"`
	Advanced     bool `json:"advanced"`
	Intermediate bool `json:"intermediate"`
	Go           bool `json:"go"`
	Docker       bool `json:"docker"`
	GCP          bool `json:"gcp"`
	Java         bool `json:"java"`
	Vue          bool `json:"vue"`
	Serverless   bool `json:"serverless"`
	AMQP         bool `json:"amqp"`
	Databases    bool `json:"databases"`
	Mux          bool `json:"mux"`
	GRPC         bool `json:"grpc"`
	Spring       bool `json:"spring"`
}

type Request struct {
	Params struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	} `json:"params"`
	Filter Filter `json:"data"`
}

type envelope map[string]any

func (c *Courses) CoursesAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	clog := log.GetLoggerFromContext(ctx)

	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		clog.ErrorCtx(ErrBadRequestBody, log.Ctx{
			"error": err,
		})

		http.Error(w, ErrBadRequestBody.Error(), http.StatusBadRequest)

		return
	}

	pipeline := make([]bson.M, 3)

	page := req.Params.Page
	pageSize := req.Params.PageSize

	// Pagination
	page = (page - 1) * pageSize

	// Add filter
	matches, err := c.getFilter(&req.Filter)
	if err != nil {
		clog.ErrorCtx(ErrGetFilter, log.Ctx{
			"error": err,
		})

		w.WriteHeader(http.StatusBadRequest)
	}

	pipeline[0] = bson.M{"$skip": page}
	pipeline[1] = bson.M{"$limit": pageSize}
	pipeline[2] = bson.M{"$match": matches}

	if !(pageSize > 0) {
		http.Error(w, "invalid pagination request", http.StatusBadRequest)

		return
	}

	res, err := c.repo.GetAll(ctx, pipeline)
	if err != nil {
		clog.ErrorCtx(ErrDecodingCourse, log.Ctx{
			"error": err,
		})

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	err = c.helper.WriteJSON(w, http.StatusOK, envelope{"data": res, "metadata": "none"}, nil)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{
			"header":      w.Header(),
			"request_url": r.URL.String(),
			"msg":         "unable to get courses",
		})
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (c *Courses) getFilter(filter *Filter) (bson.M, error) {

	match := bson.M{}
	tags := []string{}

	if filter.All {
		// Include all items
		tags = append(tags, "beginner")
		tags = append(tags, "intermediate")
		tags = append(tags, "advanced")
		tags = append(tags, "go")
		tags = append(tags, "docker")
		tags = append(tags, "gcp")
		tags = append(tags, "java")
		tags = append(tags, "vue")
		tags = append(tags, "serverless")
		tags = append(tags, "amqp")
		tags = append(tags, "databases")
		tags = append(tags, "mux")
		tags = append(tags, "grpc")
		tags = append(tags, "spring")
	} else {
		// Filter items based on the provided criteria
		if filter.Beginner {
			tags = append(tags, "beginner")
		}
		if filter.Advanced {
			tags = append(tags, "advanced")
		}
		if filter.Go {
			tags = append(tags, "go")
		}
		if filter.Docker {
			tags = append(tags, "docker")
		}
		if filter.GCP {
			tags = append(tags, "gcp")
		}
		if filter.Java {
			tags = append(tags, "java")
		}
		if filter.Vue {
			tags = append(tags, "vue")
		}
		if filter.Serverless {
			tags = append(tags, "serverless")
		}
		if filter.AMQP {
			tags = append(tags, "amqp")
		}
		if filter.Databases {
			tags = append(tags, "databases")
		}
		if filter.Mux {
			tags = append(tags, "mux")
		}
		if filter.GRPC {
			tags = append(tags, "grpc")
		}
		if filter.Spring {
			tags = append(tags, "spring")
		}

	}

	match = bson.M{
		"topic": bson.M{"$in": tags},
	}

	return match, nil
}

func (c *Courses) CoursesIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	clog := log.GetLoggerFromContext(ctx)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		clog.ErrorCtx(ErrBadRequestBody, log.Ctx{
			"error": ErrMissingID,
		})

		json.NewEncoder(w).Encode(ErrResponse{
			Message: ErrBadRequestBody,
		})

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	res, err := c.repo.GetCourseById(ctx, id)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{
			"error": ErrUnableToGetCourse,
		})

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(ErrResponse{
			Message: ErrUnableToGetCourse,
		})

		return
	}

	err = c.helper.WriteJSON(w, http.StatusOK, envelope{"data": res, "metadata": "none"}, nil)
	if err != nil {
		log.ErrorCtx(err, log.Ctx{
			"header":      w.Header(),
			"request_url": r.URL.String(),
		})
	}
}

func (c *Courses) CreateCourseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	clog := log.GetLoggerFromContext(ctx)

	var course Course

	err := json.NewDecoder(r.Body).Decode(&course)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{
			"msg": ErrDecodingCourse,
		})

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(ErrDecodingCourse)

		return
	}

	validateCourse(&c.validator, &course)

	if !c.validator.Valid() {
		clog.ErrorCtx(err, log.Ctx{
			"msg": ErrInvalidCourseFormat,
		})

		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(ErrInvalidCourseFormat)

		return
	}

	err = c.repo.CreateCourse(ctx, &course)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{
			"msg": ErrCreateCourse,
		})

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(ErrCreateCourse)

		return
	}

	err = c.helper.WriteJSON(w, http.StatusOK, envelope{"data": 200, "metadata": "none"}, nil)
	if err != nil {
		log.ErrorCtx(err, log.Ctx{
			"header":      w.Header(),
			"request_url": r.URL.String(),
		})
	}

}

func (c *Courses) UpdateCourseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	clog := log.GetLoggerFromContext(ctx)

	var cc Course

	err := c.helper.ReadJSON(w, r, &c)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{
			"msg": ErrDecodingCourse,
		})
	}

	clog.InfoCtx("course object", log.Ctx{
		"lessons":     cc.Lessons,
		"title":       cc.Title,
		"child title": cc.Details.Title,
	})

	//TODO: add a validator here for course

	err = c.repo.UpdateCourse(ctx, &cc)
	if err != nil {
		clog.Error(err)
	}

}

func (c *Courses) DeleteCourseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	clog := log.GetLoggerFromContext(ctx)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		clog.ErrorCtx(ErrBadRequestBody, log.Ctx{
			"error": ErrMissingID,
		})

		json.NewEncoder(w).Encode(ErrResponse{
			Message: ErrBadRequestBody,
		})

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	err := c.repo.DeleteCourse(ctx, id)
	if err != nil {
		clog.ErrorCtx(err, log.Ctx{
			"msg": ErrDeleteCourse,
		})

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)

}

func validateCourse(v *validator.Validator, course *Course) {
	v.Check(course.Topic != "", "topic", "topic must be provided")
	v.Check(course.Title != "", "title", "title must be provided")
}

func NewCoursesService(repo Repository) *Courses {
	return &Courses{
		repo:      repo,
		validator: *validator.New(),
	}
}
