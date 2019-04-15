package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGauge(t *testing.T) {
	t.Run("initialize new Gauge", func(t *testing.T) {
		g := NewGauge()
		assert.Equal(t, 0.0, g.Value(), "initial value should be 0.0")
	})

	t.Run("get gauge value", func(t *testing.T) {

	})
}
