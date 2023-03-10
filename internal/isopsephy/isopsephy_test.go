package isopsephy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculate(t *testing.T) {
	cases := []struct {
		Name   string
		Number int
		Error  bool
	}{
		{"Freddy", 0, true},
		{"Ἀφροδίτη", 993, false},
		{"Ͷιχανους", 1187, false},
	}

	assert := assert.New(t)

	for _, c := range cases {
		actual, err := Calculate(c.Name)
		if c.Error {
			assert.Error(err)
			continue
		}

		if !assert.NoError(err) {
			continue
		}

		assert.Equal(c.Number, actual)
	}
}
