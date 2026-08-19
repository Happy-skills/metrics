package repository

import (
	"path"
	"testing"

	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsInFile(t *testing.T) {
	tempStorage := path.Join(t.TempDir(), "store.json")

	t.Run("WriteMetricsInFile", func(t *testing.T) {
		outputStore := NewMemStorage("", 0)

		err := outputStore.SetValue(t.Context(), models.Gauge, "test", 1234)
		require.NoError(t, err)

		err = WriteMetricsInFile(tempStorage, outputStore)
		require.NoError(t, err)
	})

	t.Run("ReadMetricsInFile", func(t *testing.T) {
		inputStore := NewMemStorage("", 0)
		LoadMetricsFromFile(tempStorage, inputStore)

		metric, err := inputStore.GetValue(t.Context(), models.Gauge, "test")
		require.NoError(t, err)

		assert.Equal(t, float64(1234), *metric.Value)
		assert.Equal(t, "test", metric.ID)
		assert.Equal(t, models.Gauge, metric.MType)
	})
}
