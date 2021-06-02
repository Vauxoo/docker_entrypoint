package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadLines(t *testing.T) {
	expected := []string{"1", "2", "a", "b", "c d f", "1 4 5"}
	res, err := readLines("testdata/somelines")
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestAppendFiles(t *testing.T) {
	expected := []string{"0", "1", "2", "3", "4", "5", "6"}
	err := appendFiles("testdata/odoo_cfg", "testdata/odoo.d")
	assert.NoError(t, err)
	res, err := readLines("testdata/odoo_cfg")
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}
