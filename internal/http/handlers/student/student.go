package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/pixel2packet/studentAPI/internal/storage"
	"github.com/pixel2packet/studentAPI/internal/types"
	"github.com/pixel2packet/studentAPI/internal/utils/response"
)

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("creating a student")

		var student types.Student

		err := json.NewDecoder(r.Body).Decode(&student)
		/*
			json.NewDecoder(r.Body).Decode(&student)
			→ Decodes the incoming request body (JSON) into the student struct.

			If the body is empty:
				→ io.EOF is returned.
				→ Respond with 400 Bad Request and a message saying "empty body".
		*/
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			/*
				If decoding failed for other reasons (e.g., invalid JSON format)
				→ Respond with a 400 Bad Request and the actual error message.
			*/
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// ✅ Validate Request Payload
		if err := validator.New().Struct(student); err != nil {
			/*
				validator.New()
					→ Creates a validator instance from the go-playground/validator/v10 package.

				.Struct(student)
					→ Validates the struct using its `validate:"..."` tags.
					→ If any field fails validation (e.g., required, email format, etc.), it returns an error.
			*/
			validateErrs := err.(validator.ValidationErrors)
			/*
				Type assertion: ensures we can format the validation errors properly for the client.
			*/
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		lastId, err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)
		/*
			Calls the storage layer to insert the student into the DB.
			Returns the last inserted row ID, or an error.
		*/

		if err != nil {
			/*
				❌ If database insert fails → respond with 500 Internal Server Error.
			*/
			response.WriteJson(w, http.StatusInternalServerError, err)
			return
		}

		slog.Info("user created successfully", slog.String("userId", fmt.Sprint(lastId)))
		/*
			Log success after DB insert — now we know the operation succeeded
		*/

		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})
		/*
			Respond with:
				→ HTTP 201 Created
				→ JSON: {"id": 123}
		*/
	}
}


func GetById(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")
		/*
			Retrieves the "id" from the request URL path.
			Example: If request is GET /api/students/42 → id = "42" (string)
		*/

		slog.Info("getting a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		/*
			Converts the string `id` to int64.
			Base 10 (decimal), 64-bit integer.
			If invalid (e.g., "abc"), respond with 400 Bad Request.
		*/
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		student, err := storage.GetStudentById(intId)
		/*
			Queries the database for a student with the given ID.
			Returns the student struct or an error.
		*/

		if err != nil {
			slog.Error("error getting user", slog.String("id", id))
			/*
				Logs the failure, then sends a 500 Internal Server Error.
			*/
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, student)
		/*
			Returns 200 OK and the student object in JSON format.
		*/
	}
}


func GetList(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		slog.Info("getting all students")
		/*
			Logs the request for fetching all students.
		*/

		students, err := storage.GetStudents()
		/*
			Calls the storage layer to retrieve all students from the DB.
			Returns a slice of student structs or an error.
		*/

		if err != nil {
			/*
				If the DB call fails → send a 500 Internal Server Error.
			*/
			response.WriteJson(w, http.StatusInternalServerError, err)
			return
		}

		response.WriteJson(w, http.StatusOK, students)
		/*
			On success:
				→ Respond with 200 OK
				→ Send all students in JSON format as a slice
		*/
	}
}

