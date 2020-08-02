package climacell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetWeather(t *testing.T) {
	weather, err := CurrentConditions("austin tx")
	require.NoError(t, err)
	t.Fatalf("%s", weather)
}
