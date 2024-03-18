package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewBloomRing(t *testing.T) {
	t.Parallel()
	br := NewBloomRing()
	assert.Equal(t, 0, br.entryCounter)
	assert.Equal(t, 100000, br.slotCapacity)
	assert.Equal(t, 10, br.slotCount)
	assert.Equal(t, 0, br.slotPosition)
	assert.Len(t, br.slots, 10)
}

func Test_BloomRing_AddSalt(t *testing.T) {
	t.Parallel()
	t.Run("first salt", func(t *testing.T) {
		t.Parallel()
		br := NewBloomRing()
		br.AddSalt(make([]byte, 16))
		assert.Equal(t, 1, br.entryCounter)
	})
	t.Run("reset salts", func(t *testing.T) {
		t.Parallel()
		br := NewBloomRing()
		br.entryCounter = br.slotCapacity + 1
		br.AddSalt(make([]byte, 16))
		assert.Equal(t, 1, br.entryCounter)
	})
}

func Test_BloomRing(t *testing.T) {
	t.Parallel()
	br := NewBloomRing()
	br.AddSalt([]byte("repeating one"))
	assert.True(t, br.IsSaltRepeated([]byte("repeating one")))
	assert.False(t, br.IsSaltRepeated([]byte("new one")))
}
