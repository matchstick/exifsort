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

// Add Some filepaths to the index
// File names are of the form "IMG_<start>.jpg" to "IMG_<end>.jpg"
// We have a time associated with each file based on args provided.
func PutFiles(t *testing.T, idx index, start uint, count uint,
		year int, month int, day int) {
	for ii := start; ii < start+count; ii++ {
		time := time.Date(year, time.Month(month), day,
				0, 0, 0, 0, time.Local)
		name := fmt.Sprintf("IMG_%d.jpg", ii)
		idx.Put(name, time)
	}
}

func GetFiles(t *testing.T, idx index, start uint, count uint,
		year int, month int, day int) {
	for ii := start; ii < start+count; ii++ {
		testTime := time.Date(year, time.Month(month), day,
				0, 0, 0, 0, time.Local)
		name := fmt.Sprintf("IMG_%d.jpg", ii)
		idxPath, present := idx.Get(name)
		if present == false {
			t.Fatalf("Cannot find %s\n", name)
		}
		testPath := idx.NewPath(testTime, name)
		if idxPath != testPath {
			t.Fatalf("Recvd paths (%s) != %s using time %s\n",idxPath, testPath, testTime)
		}
	}
}

func TestYearIndex(t *testing.T) {
	var idx = CreateIndex(METHOD_YEAR)
	PutFiles(t, idx, 10, 10, 2020, 1, 1)
	GetFiles(t, idx, 10, 10, 2020, 1, 1)
	fmt.Printf("%s\n", idx)
}
func TestMonthIndex(t *testing.T) {
	var idx = CreateIndex(METHOD_MONTH)
	PutFiles(t, idx, 10, 10, 1, 2, 1)
	GetFiles(t, idx, 10, 10, 1, 2, 1)
	fmt.Printf("%s\n", idx)
}

func TestDayIndex(t *testing.T) {
	var idx = CreateIndex(METHOD_DAY)
	PutFiles(t, idx, 10, 10, 1, 2, 4)
	GetFiles(t, idx, 10, 10, 1, 2, 4)
	fmt.Printf("%s\n", idx)
}
