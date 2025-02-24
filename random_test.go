package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	length := 10
	r1 := randomString(length)
	r2 := randomString(length)

	assert.Equal(t, length, len(r1), "expected length to be %d", length)
	assert.Equal(t, length, len(r2), "expected length to be %d", length)
	assert.NotEqual(t, r1, r2, "expected %s to not be equal to %s", r1, r2)
}

func TestRandomStringDifferentLengths(t *testing.T) {
	lengths := []int{5, 10, 15, 20}
	for _, length := range lengths {
		t.Run(fmt.Sprintf("%d", length), func(t *testing.T) {
			result := randomString(length)
			assert.Equal(t, length, len(result), "expected length to be %d", length)
		})
	}
}

func TestRandomStringZeroLength(t *testing.T) {
	result := randomString(0)
	assert.Equal(t, 0, len(result), "expected length to be 0")
}

func TestRandomStringNegativeLength(t *testing.T) {
	result := randomString(-1)
	assert.Equal(t, 0, len(result), "expected length to be 0")
}
