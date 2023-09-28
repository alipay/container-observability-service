package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertNil(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", "Nil"},
		{"Non-empty string", "value", "value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ConvertNil(tt.input))
		})
	}
}

func TestExpiringMap(t *testing.T) {
	em := NewExpiringMap(100 * time.Millisecond)

	// Add a key-value pair and check if it exists
	em.Set("foo", "bar")
	val, found := em.Get("foo")
	assert.True(t, found)
	assert.Equal(t, "bar", val)

	// Delete the key-value pair and check if it no longer exists
	em.Delete("foo")
	_, found = em.Get("foo")
	assert.False(t, found)

	// Wait for timeout and check if the key-value pair is deleted
	em.Set("foo", "bar")
	time.Sleep(200 * time.Millisecond)
	_, found = em.Get("foo")
	assert.False(t, found)
}
