package climacell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testIsValidFeature(t *testing.T) {
	assert.True(t, isValidFeature("temp"))
	assert.True(t, isValidFeature("precipitation"))
	assert.True(t, isValidFeature("temp"))
	assert.True(t, isValidFeature("wind_speed"))
	assert.True(t, isValidFeature("wind_direction"))
	assert.True(t, isValidFeature("wind_gust"))
	assert.True(t, isValidFeature("visibility"))
	assert.True(t, isValidFeature("baro_pressure"))
	assert.True(t, isValidFeature("dewpoint"))
	assert.True(t, isValidFeature("humidity"))
	assert.True(t, isValidFeature("cloud_cover"))
	assert.True(t, isValidFeature("cloud_base"))
	assert.True(t, isValidFeature("cloud_ceiling"))
	assert.True(t, isValidFeature("cloud_satellite"))
	assert.False(t, isValidFeature("foo"))
}
