package utils_test

import (
	"testing"

	"github.com/niyoko/family-assistant/src/utils"
	"github.com/stretchr/testify/assert"
)

func TestTrimStringToMaxLen(t *testing.T) {
	assert.Equal(t, "123", utils.TrimStringToMaxLen("123", 3))
	assert.Equal(t, "123", utils.TrimStringToMaxLen("123", 4))
	assert.Equal(t, "日", utils.TrimStringToMaxLen("日本語", 5))
	assert.Equal(t, "日本", utils.TrimStringToMaxLen("日本語", 6))
	assert.Equal(t, "日本", utils.TrimStringToMaxLen("日本語", 7))
	assert.Equal(t, "日本", utils.TrimStringToMaxLen("日本語", 8))
	assert.Equal(t, "日本語", utils.TrimStringToMaxLen("日本語", 9))
	assert.Equal(t, "日本語", utils.TrimStringToMaxLen("日本語", 10))
}

func TestChunkLinesToMaxLen(t *testing.T) {
	assert.Equal(t, [][]string{
		{"contoh pa"},
	}, utils.ChunkLinesToMaxLen([]string{
		"contoh paragraf yang panjang",
	}, 10))

	assert.Equal(t, [][]string{
		{"contoh par"},
	}, utils.ChunkLinesToMaxLen([]string{
		"contoh paragraf yang panjang",
	}, 11))

	assert.Equal(t, [][]string{
		{"This library is v1 and follows SemVer strictly."},
		{"No breaking changes will be made to exported APIs before v2.0.0."},
		{"This library has no dependencies outside the Go standard library."},
	}, utils.ChunkLinesToMaxLen([]string{
		"This library is v1 and follows SemVer strictly.",
		"No breaking changes will be made to exported APIs before v2.0.0.",
		"This library has no dependencies outside the Go standard library.",
	}, 100))

	assert.Equal(t, [][]string{
		{
			"This library is v1 and follows SemVer strictly.",
			"No breaking changes will be made to exported APIs before v2.0.0.",
		},
		{
			"This library has no dependencies outside the Go standard library.",
		},
	}, utils.ChunkLinesToMaxLen([]string{
		"This library is v1 and follows SemVer strictly.",
		"No breaking changes will be made to exported APIs before v2.0.0.",
		"This library has no dependencies outside the Go standard library.",
	}, 150))

	assert.Equal(t, [][]string{
		{
			"This library is v1 and follows SemVer strictly.",
			"No breaking changes will be made to exported APIs before v2.0.0.",
			"This library has no dependencies outside the Go standard library.",
		},
	}, utils.ChunkLinesToMaxLen([]string{
		"This library is v1 and follows SemVer strictly.",
		"No breaking changes will be made to exported APIs before v2.0.0.",
		"This library has no dependencies outside the Go standard library.",
	}, 200))
}
