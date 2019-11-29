package nws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeather(t *testing.T) {
	assert := assert.New(t)
	r, err := GetWeather("78703")
	assert.NoError(err)
	assert.NotNil(r)
	assert.NotEqual("", r.String())
	t.Logf("%s", r.String())
}
