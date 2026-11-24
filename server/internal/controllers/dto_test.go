package controllers

import (
	"encoding/json"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestamp_MarshalJSON(t *testing.T) {
	central := time.FixedZone("", -6*60*60)
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"whole seconds", time.Date(2026, 11, 24, 13, 54, 18, 0, time.UTC), `"2026-11-24T13:54:18Z"`},
		{"finer than seconds is cut", time.Date(2026, 11, 24, 13, 54, 18, 999_999_999, time.UTC), `"2026-11-24T13:54:18Z"`},
		{"other zones are sent as UTC", time.Date(2026, 11, 24, 13, 54, 18, 0, central), `"2026-11-24T19:54:18Z"`},
	}

	// the pattern the OpenAPI spec promises
	field, ok := reflect.TypeFor[meDTO]().FieldByName("CreatedAt")
	require.True(t, ok)
	pattern := regexp.MustCompile(field.Tag.Get("pattern"))

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(timestamp(tc.time))
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(got))

			var s string
			require.NoError(t, json.Unmarshal(got, &s))
			assert.Regexp(t, pattern, s)
			parsed, err := time.Parse(time.RFC3339, s)
			require.NoError(t, err)
			assert.True(t, parsed.Equal(tc.time.Truncate(time.Second)))
		})
	}
}
