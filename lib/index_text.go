package exifSort

import (
	"fmt"
	//"errors"
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
// 
func AddFilesByYear(t *testing.T, idx *index,
			start uint, count uint, year int) {
	for ii := start; ii < start+count; ii++ {
		time := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		name := fmt.Sprintf("IMG_%d.jpg", ii)
		idx.AddMediaByYear(name, time)
	}
}

func TestSortedTimeAddByYear(t *testing.T) {
	var idx index
	idx.InitRoot(METHOD_YEAR)
	// Populate the sortedType
	for ii:=2000; ii<2020; ii+=2 {
		AddFilesByYear(t, &idx, 10, 10, ii)
		AddFilesByYear(t, &idx, 15, 5, ii)
	}
	fmt.Printf("%s\n", idx)
}


func TestSortedTimeAddByMonth(t *testing.T) {
}


