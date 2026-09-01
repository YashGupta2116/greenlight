package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"greenlight.codewithyash.dev/internal/validator"
)

type envelope map[string]any


func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}


func (app *application) writeJSON(w http.ResponseWriter, status int,  data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field !="" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return  fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(),"json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(),"json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return  strings.Split(s, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	value := qs.Get(key)

	if value == "" {
		return defaultValue
	}

	convertedVal, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return convertedVal
}


/* More readable way
func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func ()  {
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		} ()

		fn()
	} ()
}
*/


/* Modern way using waitgroup.go */
func (app *application) background(fn func()) {
	app.wg.Go(func() {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		} ()

		fn()
	})
}


/* Config setup using env */
func getEnv(key string) string {
	return os.Getenv(key)
}

func (cfg *config) getConfiguration() error {
	var err error

	cfg.env = getEnv("ENV")

	cfg.port, err = strconv.Atoi(getEnv("PORT"))
	if err != nil {
		return err
	}

	cfg.db.dsn = getEnv("DB_DSN")

	cfg.db.maxOpenConns, err = strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS"))
	if err != nil {
		return err
	}

	cfg.db.maxIdleConns, err = strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS"))
	if err != nil {
		return err
	}

	cfg.db.maxIdleTime, err = time.ParseDuration(getEnv("DB_MAX_IDLE_TIME"))
	if err != nil {
		return err
	}

	cfg.limiter.rps, err = strconv.ParseFloat(getEnv("LIMITER_RPS"), 64)
	if err != nil {
		return err
	}

	cfg.limiter.burst, err = strconv.Atoi(getEnv("LIMITER_BURST"))
	if err != nil {
		return err
	}

	cfg.limiter.enabled, err = strconv.ParseBool(getEnv("LIMITER_ENABLED"))
	if err != nil {
		return err
	}

	cfg.smtp.host = getEnv("SMTP_HOST")

	cfg.smtp.port, err = strconv.Atoi(getEnv("SMTP_PORT"))
	if err != nil {
		return err
	}

	cfg.smtp.username = getEnv("SMTP_USERNAME")
	cfg.smtp.password = getEnv("SMTP_PASSWORD")
	cfg.smtp.sender = getEnv("SMTP_SENDER")

	return nil
}
