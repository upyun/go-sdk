package upyun

import "testing"

func TestSkipter(t *testing.T) {
	skiper := createSkiper(
		2,
		[]*DisorderPart{
			&DisorderPart{ID: 0, Size: 1},
			&DisorderPart{ID: 1, Size: 1},
			&DisorderPart{ID: 3, Size: 1},
			&DisorderPart{ID: 5, Size: 1},
			&DisorderPart{ID: 10, Size: 1},
			&DisorderPart{ID: 11, Size: 1},
		},
	)

	Equal(t, skiper.FirstMissPartId(), int64(2))
	Equal(t, skiper.IsSkip(4), false)
	Equal(t, skiper.IsSkip(5), true)
	Equal(t, skiper.IsSkip(6), false)
	Equal(t, skiper.IsSkip(7), false)
	Equal(t, skiper.IsSkip(8), false)
	Equal(t, skiper.IsSkip(9), false)
	Equal(t, skiper.IsSkip(10), true)
}
