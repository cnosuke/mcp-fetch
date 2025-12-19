package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrap(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrap(baseErr, "context message")

	require.NotNil(t, wrapped)
	assert.Contains(t, wrapped.Error(), "context message")
	assert.Contains(t, wrapped.Error(), "base error")
	assert.Equal(t, "context message: base error", wrapped.Error())
}

func TestWrap_NilError(t *testing.T) {
	wrapped := Wrap(nil, "context message")
	assert.Nil(t, wrapped)
}

func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrapf(baseErr, "context with %s and %d", "string", 42)

	require.NotNil(t, wrapped)
	assert.Contains(t, wrapped.Error(), "context with string and 42")
	assert.Contains(t, wrapped.Error(), "base error")
	assert.Equal(t, "context with string and 42: base error", wrapped.Error())
}

func TestWrapf_NilError(t *testing.T) {
	wrapped := Wrapf(nil, "context with %s", "value")
	assert.Nil(t, wrapped)
}

func TestWrap_ErrorChain(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrap(baseErr, "context message")

	// errors.Is should work with error chain
	assert.True(t, errors.Is(wrapped, baseErr))
}

func TestWrapf_ErrorChain(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrapf(baseErr, "context with %s", "value")

	// errors.Is should work with error chain
	assert.True(t, errors.Is(wrapped, baseErr))
}
