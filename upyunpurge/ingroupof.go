package upyunpurge

// Separate a slice into sub-slices with length of n, except for
// the last sub-slice, whose length may be less than n.
// For example:
//   InGroupOf([]{ "a", "b", "c", "d", "e"}, 2) =
//      [][]string{{"a", "b"}, {"c", "d"}, {"e"}}
func InGroupOf(items []string, n int) [][]string {
	size := len(items)

	division := size / n
	modulo := size % n

	newSize := division
	if modulo > 0 {
		newSize += 1
	}

	list := make([][]string, newSize)
	for i := 0; i < newSize; i++ {
		start := i * n
		end := (i + 1) * n
		if end > size {
			end = size
		}
		list[i] = items[start:end]
	}

	return list
}

// Merge sub-slices into a slice.
// For example:
//   ConcatSubGroups([][]string{ {"a", "b"}, {"c"}}) =
//      []String{"a", "b", "c"}
func ConcatSubGroups(listOfList [][]string) []string {
	finalList := make([]string, 0)
	for _, items := range listOfList {
		finalList = append(finalList, items...)
	}

	return finalList
}
