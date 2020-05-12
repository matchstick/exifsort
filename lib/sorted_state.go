package exifSort

import (
	"fmt"
	"time"
)

const (
	SORT_NONE = iota
	SORT_YEAR
	SORT_MONTH
	SORT_DAY
)


// If the value matches key then it is a file that is not being moved.
type mediaMap map[string]string

// handle collisions
func (m *mediaMap) add(path string) {
// TODO handle collisions
}

/*
type DayBucket struct {
	Media mediaMap
	day int
}

func (d *DayBucket) Day() int {
	return month
}


type MonthBucket struct {
	Media mediaMap
	days  DayBucket[]
	month time.Month
}

func (m *MonthBucket) AddDay(day int) {

}

func (m *MonthBucket) Month() time.Month {
	return m.month
}
*/
type YearBucket struct {
	media  mediaMap
//	months []MonthBucket
	year   int
}
/*
func (y *YearBucket) AddMonth(month time.Month) {
	newMonth := New(MonthBucket)
	newMonth.month = month
}
*/

func (y *YearBucket) Year() int {
	return y.year
}

type SortedType struct {
	years  map[int]YearBucket
	method int
	root   string
}

func (st *SortedType) AddByYear(path string, time time.Time) {
	year := time.Year()
	yearBucket, present := st.years[year]
	if present == false {
		yearBucket = YearBucket{nil, year}
		st.years[year] = yearBucket
	}
	yearBucket.media.add(path)
}

func (st *SortedType) Init(root string, method int) {
	st.years = nil
	st.method = method
	st.root = root
}

func (st *SortedType) Add(path string, time time.Time) {

	switch st.method {
	case SORT_YEAR:
		st.AddByYear(path, time)
	}
	fmt.Printf("Sorting not readyyet\n")
}

func (st *SortedType) String() string {
	var retStr string
	for _, year := range st.years {
		retStr += fmt.Sprintf("%d \n", year.Year())
	}
	return retStr
}
