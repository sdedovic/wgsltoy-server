package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

func TestWriteError_JSON(t *testing.T) {
	tests := []struct {
		err    error
		status int
		class  string
	}{
		{domain.NewValidationError("Field 'name' may not be empty!"), http.StatusUnprocessableEntity, "VALIDATION_FAILURE"},
		{domain.ErrUsernameTaken, http.StatusConflict, "CONFLICT"},
		{domain.ErrEmailTaken, http.StatusConflict, "CONFLICT"},
		{domain.ErrBadLogin, http.StatusUnauthorized, "BAD_LOGIN"},
		{domain.ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED"},
		{domain.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{ErrUnsupportedMediaType, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE"},
		{ErrNotAcceptable, http.StatusNotAcceptable, "NOT_ACCEPTABLE"},
		{errors.New("database is down"), http.StatusInternalServerError, "UNKNOWN"},
	}
	for _, tc := range tests {
		t.Run(tc.class, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()

			WriteError(w, r, tc.err)

			assert.Equal(t, tc.status, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			assert.Contains(t, w.Body.String(), `"errorClass":"`+tc.class+`"`)
		})
	}
}

func TestWriteError_HTML(t *testing.T) {
	t.Run("unauthorized goes to login", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteError(w, httptest.NewRequest("GET", "/me", nil), domain.ErrUnauthorized)
		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/login", w.Header().Get("Location"))
	})

	t.Run("validation shows its message", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteError(w, httptest.NewRequest("POST", "/shaders", nil), domain.NewValidationError("Field 'name' may not be empty!"))
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Contains(t, w.Body.String(), "Field 'name' may not be empty!")
	})

	t.Run("server errors hide their cause", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteError(w, httptest.NewRequest("GET", "/", nil), errors.New("database is down"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "database")
	})
}

func TestHandle_VariesOnAccept(t *testing.T) {
	w := httptest.NewRecorder()
	Handle(func(w http.ResponseWriter, r *http.Request) error { return nil }).
		ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, "Accept", w.Header().Get("Vary"))
}
