package climacell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetWeather(t *testing.T) {
	weather, image, err := CurrentConditions("austin tx")
	require.NoError(t, err)
	t.Fatalf("%s", weather)
	require.NotNil(t, image)
}
