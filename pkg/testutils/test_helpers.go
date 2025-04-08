package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHelper provides common testing utilities
type TestHelper struct {
    T *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
    return &TestHelper{T: t}
}

func (h *TestHelper) AssertEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
    assert.Equal(h.T, expected, actual, msgAndArgs...)
}

func (h *TestHelper) AssertNoError(err error, msgAndArgs ...interface{}) {
    assert.NoError(h.T, err, msgAndArgs...)
}