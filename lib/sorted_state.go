package exifSort

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	METHOD_NONE = iota
	METHOD_YEAR
	METHOD_MONTH
	METHOD_DAY
)

var methodMap = map[int]string{
	METHOD_NONE:  "None",
	METHOD_YEAR:  "Year",
	METHOD_MONTH: "Month",
	METHOD_DAY:   "Day",
}

func methodStr(method int) string {
	str, present := methodMap[method]
	if present == false {
		return "unknown"
	}
	return str
}

// key   == base of path
// value == original path
type mediaMap map[string]string

func mediaCollisionName(m mediaMap, base string) string {
	var name string
	var newName string
	pieces := strings.Split(base, ".")
	numPieces := len(pieces)
	// get the suffx
	suffix := pieces[numPieces-1]
	// reconstruct the name (have to handle multiple "." in name)
	for i := 0; i < numPieces-1; i++ {
		name += pieces[i]
	}
	// Now we keep trying until we create a name that won't collide
	for counter := 0; true; counter++ {
		newName = fmt.Sprintf("%s_%d.%s", name, counter, suffix)
		_, present := m[newName]
		if present == false {
			break
		}
	}
	return newName
}

// handle collisions
func mediaMapAdd(m mediaMap, path string) mediaMap {
	var base = filepath.Base(path)
	_, present := m[base]
	if present {
		base = mediaCollisionName(m, base)
	}
	m[base] = path
	return m
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
	media mediaMap
	//	months []MonthBucket
	year int
}

func (y *YearBucket) Init(year int) {
	y.media = make(mediaMap)
	y.year = year
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

type sortedType struct {
	years  map[int]YearBucket
	method int
	root   string
}

func (st *sortedType) AddByYear(path string, time time.Time) {
	year := time.Year()
	yearBucket, present := st.years[year]
	if present == false {
		yearBucket.Init(year)
		st.years[year] = yearBucket
	}
	yearBucket.media = mediaMapAdd(yearBucket.media, path)
}

func (st *sortedType) Init(root string, method int) {
	st.years = make(map[int]YearBucket)
	st.method = method
	st.root = root
}

func (st *sortedType) Add(path string, time time.Time) {
	switch st.method {
	case METHOD_YEAR:
		st.AddByYear(path, time)
	}
}

func (st sortedType) String() string {
	var retStr string
	retStr += fmt.Sprintf("Root: %s\n", st.root)
	retStr += fmt.Sprintf("Method: By %s\n", methodStr(st.method))
	for _, year := range st.years {
		retStr += fmt.Sprintf("Year: %d\n", year.Year())
		for base, srcPath := range year.media {
			retStr += fmt.Sprintf("\t%s=>\t%s\n", srcPath, base)
		}
	}
	return retStr
}
