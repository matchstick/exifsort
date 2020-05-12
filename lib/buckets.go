package exifSort

import (
	"time"
)

// If the value matches key then it is a file that is not being moved.
type mediaMap string[string]

// handle collisions
func (m *mediaMap) addMedia(path string) error {
// TODO handle collisions
}


type DayBucket struct {
	Media mediaMap
	day int
}

func (d DayBucket) Day() int {
	return month
}


type MonthBucket struct {
	Media mediaMap
	days  DayBucket[]
	month time.Month
}

func (m MonthBucket) AddDay(day int) {

}

func (m MonthBucket) Month() time.Month {
	return m.month
}

type YearBucket struct {
	Media  mediaMap
	months MonthBucket[]
	year   int
}

func (y YearBucket) AddMonth(month time.Month) {
	newMonth := New(MonthBucket)
	newMonth.month = month
	months.
}

func (y YearBucket) Year() int {
	return y.month
}



