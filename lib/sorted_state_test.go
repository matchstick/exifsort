package exifSort

import (
	"fmt"
	//"errors"
	"testing"
	"time"
)

var ByYearInput = []string{
	"b/a.jpg",
	"c/c.jpg",
	"d/c.jpg",
	"d/c.jpg",
	"d/c.jpg",
}

func TestSortedTimeAddByYear(t *testing.T) {
	var st sortedType
	st.Init("root", SORT_YEAR)
	time := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, path := range ByYearInput {
		st.Add(path, time)
	}
	fmt.Printf("%s", st)
}
