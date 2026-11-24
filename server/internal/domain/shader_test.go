package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewShaderContent_Length(t *testing.T) {
	line := "  let color = 0.5 + 0.5 * cos(iTime + uv.xyx + vec3f(0, 2, 4));\n"

	_, err := NewShaderContent(strings.Repeat(line, 400))
	assert.NoError(t, err, "a 400 line shader")

	_, err = NewShaderContent(strings.Repeat("a", maxShaderContentLength))
	assert.NoError(t, err)

	_, err = NewShaderContent(strings.Repeat("a", maxShaderContentLength+1))
	assert.IsType(t, ValidationError{}, err)
}
