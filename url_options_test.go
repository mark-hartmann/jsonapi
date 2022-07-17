package jsonapi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestURLOptions_Path(t *testing.T) {
	// Empty
	opts := URLOptions{}

	assert.Equal(t, "/", opts.Path("/"))
	assert.Equal(t, "/type", opts.Path("/type"))
	assert.Equal(t, "/type/id", opts.Path("/type/id"))
	assert.Equal(t, "/type/id/relationship", opts.Path("/type/id/relationship"))
	assert.Equal(t, "/type/id/relationship/id", opts.Path("/type/id/relationship/id"))

	// Slash
	opts = URLOptions{Prefix: "/"}

	assert.Equal(t, "/", opts.Path("/"))
	assert.Equal(t, "/type", opts.Path("/type"))
	assert.Equal(t, "/type/id", opts.Path("/type/id"))
	assert.Equal(t, "/type/id/relationship", opts.Path("/type/id/relationship"))
	assert.Equal(t, "/type/id/relationship/id", opts.Path("/type/id/relationship/id"))
	assert.Equal(t, "/api/type/id/relationship/id", opts.Path("/api/type/id/relationship/id"))

	// Path no slash
	opts = URLOptions{Prefix: "api"}

	assert.Equal(t, "/", opts.Path("/"))
	assert.Equal(t, "/type", opts.Path("/type"))
	assert.Equal(t, "/type/id", opts.Path("/type/id"))
	assert.Equal(t, "/type/id/relationship", opts.Path("/type/id/relationship"))
	assert.Equal(t, "/type/id/relationship/id", opts.Path("/type/id/relationship/id"))
	assert.Equal(t, "/", opts.Path("/api/"))
	assert.Equal(t, "/type", opts.Path("/api/type"))
	assert.Equal(t, "/type/id", opts.Path("/api/type/id"))
	assert.Equal(t, "/type/id/relationship", opts.Path("/api/type/id/relationship"))
	assert.Equal(t, "/type/id/relationship/id", opts.Path("/api/type/id/relationship/id"))

	// Path slash
	opts = URLOptions{Prefix: "/api/"}

	assert.Equal(t, "/", opts.Path("/"))
	assert.Equal(t, "/type", opts.Path("/type"))
	assert.Equal(t, "/type/id", opts.Path("/type/id"))
	assert.Equal(t, "/type/id/relationship", opts.Path("/type/id/relationship"))
	assert.Equal(t, "/type/id/relationship/id", opts.Path("/type/id/relationship/id"))
	assert.Equal(t, "/", opts.Path("/api/"))
	assert.Equal(t, "/type", opts.Path("/api/type"))
	assert.Equal(t, "/type/id", opts.Path("/api/type/id"))
	assert.Equal(t, "/type/id/relationship", opts.Path("/api/type/id/relationship"))
	assert.Equal(t, "/type/id/relationship/id", opts.Path("/api/type/id/relationship/id"))
}
