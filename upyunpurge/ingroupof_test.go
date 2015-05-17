package upyunpurge

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestInGroupOF(t *testing.T) {
	n := 100

	numbers := make([]string, n)
	for i := 0; i < n; i++ {
		numbers[i] = strconv.Itoa(i)
	}

	for i := 1; i < n+2; i++ {
		r := InGroupOf(numbers, i)

		newNumbers := ConcatSubGroups(r)

		assert.Equal(t, len(numbers), len(newNumbers))
		for m := 0; m < len(numbers); m++ {
			assert.Equal(t, numbers[m], newNumbers[m])
		}
	}
}
