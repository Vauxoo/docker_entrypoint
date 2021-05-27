package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandStringRunes(t *testing.T) {
	res := RandStringRunes(10)
	assert.Len(t, res, 10)

	res = RandStringRunes(215)
	assert.Len(t, res, 215)
}
