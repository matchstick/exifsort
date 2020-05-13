package exifSort

import (
	"fmt"
	"testing"
	"time"
)

var ByYearInput = []string{
	"IMG.jpg",
	"c.jpg",
	"c.jpg",
	"a.jpg",
	"c.jpg",
}

// Return a a map of test filenames to times
// filenames are of the form "IMG_<start>.jpg" to "IMG_<end>.jpg"
func AddFileSet(t *testing.T, idx *index, start uint, count uint, year int, month int, day int) {
	for ii := start; ii < start+count; ii++ {
		time := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		name := fmt.Sprintf("IMG_%d.jpg", ii)
		idx.Add(name, time)
	}
}

func TestIndex(t *testing.T) {
	var idx index
	idx.InitRoot(METHOD_YEAR)
	// Populate  for a year
	for ii := 2000; ii < 2020; ii += 1 {
		AddFileSet(t, &idx, 10, 10, ii, 1, 1) // ten files  each year.
		AddFileSet(t, &idx, 15, 5, ii, 1, 1)  // five duplicate files each year
	}

	idx.InitRoot(METHOD_MONTH)
	// Populate for a month
	for ii := 1; ii <= 12; ii += 1 {
		AddFileSet(t, &idx, 10, 10, 1, ii, 1) // ten files for each month in year 1.
		AddFileSet(t, &idx, 15, 5, 1, ii, 1)  // five duplicate files each month
	}
	// Populate  for a day
	for ii := 3; ii < 20; ii += 1 {
		AddFileSet(t, &idx, 10, 10, ii, 1, 1) // ten files  each year.
		AddFileSet(t, &idx, 15, 5, ii, 1, 1)  // five duplicate files each year
	}
	fmt.Printf("%s\n", idx)
}
