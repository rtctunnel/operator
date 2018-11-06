package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannelCollection(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		c := make(chan string)
		cc := make(ChannelCollection)
		cc.Add("test", c)

		assert.NotNil(t, cc["test"])
		assert.Equal(t, c, firstkey(cc["test"]))
	})
	t.Run("Remove", func(t *testing.T) {
		c := make(chan string)
		cc := make(ChannelCollection)
		cc.Add("test", c)
		cc.Remove("test", c)
		assert.Nil(t, cc["test"])
	})
	t.Run("List", func(t *testing.T) {
		c1 := make(chan string)
		c2 := make(chan string)
		cc := make(ChannelCollection)
		cc.Add("test", c1)
		cc.Add("test", c2)
		assert.Len(t, cc.List("test"), 2)
	})
}

func firstkey(m map[chan string]struct{}) chan string {
	for c := range m {
		return c
	}
	return nil
}
