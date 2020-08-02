package geocoding

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	if ret == 0 {
		teardown()
	}
	os.Exit(ret)
}

func setup() {
	if "" == os.Getenv("GEOCODING_KEY") {
		panic("must set GEOCODING_KEY to run tests")
	}
}

func teardown() {

}

func TestGeocode(t *testing.T) {
	geocoder := new(OpenCageData)
	c, err := geocoder.Geocode("austin")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, 30.2711286, c.Latitude)
	assert.Equal(t, -97.7436995, c.Longitude)
}
