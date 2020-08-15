package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSaltFilter(t *testing.T) {
	t.Parallel()
	var saltFilter SaltFilter
	require.NotPanics(t, func() {
		saltFilter = NewSaltFilter()
	})
	require.NotNil(t, saltFilter)
}

func Test_newBloomRing(t *testing.T) {
	t.Parallel()
	br := newBloomRing()
	assert.Equal(t, 0, br.entryCounter)
	assert.Equal(t, 100000, br.slotCapacity)
	assert.Equal(t, 10, br.slotCount)
	assert.Equal(t, 0, br.slotPosition)
	assert.Equal(t, 10, len(br.slots))
}

func Test_bloomRing_AddSalt(t *testing.T) {
	t.Parallel()
	t.Run("first salt", func(t *testing.T) {
		t.Parallel()
		br := newBloomRing()
		br.AddSalt(make([]byte, 16))
		assert.Equal(t, 1, br.entryCounter)
	})
	t.Run("reset salts", func(t *testing.T) {
		t.Parallel()
		br := newBloomRing()
		br.entryCounter = br.slotCapacity + 1
		br.AddSalt(make([]byte, 16))
		assert.Equal(t, 1, br.entryCounter)
	})
}

func TestBloomRing_Test(t *testing.T) {
	t.Parallel()
	br := newBloomRing()
	br.AddSalt([]byte("repeating one"))
	assert.True(t, br.IsSaltRepeated([]byte("repeating one")))
	assert.False(t, br.IsSaltRepeated([]byte("new one")))
}
