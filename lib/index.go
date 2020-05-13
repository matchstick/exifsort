package exifSort

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// key   == i base of path
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
	for ii := 0; ii < numPieces-1; ii++ {
		name += pieces[ii]
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

func (m mediaMap) Keys() []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

type indexMap map[int]index

const ROOT_INDEX = -1

type index struct {
	media   mediaMap
	entries indexMap
	id      int
	// I hate having this switch I must fix
	method int
}

func (i *index) Media() mediaMap {
	return i.media
}

func (i *index) Init(id int) {
	i.media = make(mediaMap)
	i.entries = make(indexMap)
	i.id = id
}

func (i *index) InitRoot(method int) {
	i.media = make(mediaMap)
	i.entries = make(indexMap)
	i.id = ROOT_INDEX
	i.method = method
}

func (i *index) EntriesKeys() []int {
	var keys []int
	for k := range i.entries {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (i *index) GetIndex(id int) index {
	idx, present := i.entries[id]
	if present == false {
		idx.Init(id)
		i.entries[id] = idx
	}
	return idx
}

func (i *index) AddPath(path string) {
	i.media = mediaMapAdd(i.media, path)
}

func (i *index) AddMediaByYear(path string, time time.Time) {
	yearIndex := i.GetIndex(time.Year())
	yearIndex.AddPath(path)
}

func (i *index) AddMediaByMonth(path string, time time.Time) {
	yearIndex := i.GetIndex(time.Year())
	monthIndex := yearIndex.GetIndex(int(time.Month()))
	monthIndex.AddPath(path)
}

func (i *index) AddMediaByDay(path string, time time.Time) {
	yearIndex := i.GetIndex(time.Year())
	monthIndex := yearIndex.GetIndex(int(time.Month()))
	dayIndex := monthIndex.GetIndex(time.Day())
	dayIndex.AddPath(path)
}

// I hate these switches. Will fix with next check in
func (i *index) Add(path string, time time.Time) {
	switch i.method {
	case METHOD_YEAR:
		i.AddMediaByYear(path, time)
	case METHOD_MONTH:
		i.AddMediaByMonth(path, time)
	case METHOD_DAY:
		i.AddMediaByDay(path, time)
	default:
		panic("Unknown method")
	}
}

func (i *index) DumpByYear() string {
	var retStr string
	for _, year := range i.EntriesKeys() {
		yearIndex := i.GetIndex(year)
		media := yearIndex.Media()
		for _, base := range media.Keys() {
			retStr += fmt.Sprintf("%s => %d/%s\n",
					media[base], year, base)
		}
	}
	return retStr
}

func (i *index) DumpByMonth() string {
	var retStr string
	for _, year := range i.EntriesKeys() {
		yearIndex := i.GetIndex(year)
		for _, month := range yearIndex.EntriesKeys() {
			monthIndex := yearIndex.GetIndex(month)
			media := monthIndex.Media()
			for _, base := range media.Keys() {
				retStr += fmt.Sprintf("%s => %d/%d/%s\n",
						media[base], year, month, base)
			}
		}
	}
	return retStr
}

func (i *index) DumpByDay() string {
	var retStr string
	for _, year := range i.EntriesKeys() {
		yearIndex := i.GetIndex(year)
		for _, month := range yearIndex.EntriesKeys() {
			monthIndex := yearIndex.GetIndex(month)
			for _, day := range monthIndex.EntriesKeys() {
				dayIndex := monthIndex.GetIndex(month)
				media := dayIndex.Media()
				for _, base := range media.Keys() {
					retStr += fmt.Sprintf("%s => %d/%d/%d/%s\n",
							media[base], year, month, day, base)
				}
			}
		}
	}
	return retStr
}

// I hate these switches. Will fix with next check in
func (i index) String() string {
	switch i.method {
	case METHOD_YEAR:
		return i.DumpByYear()
	case METHOD_MONTH:
		return i.DumpByMonth()
	case METHOD_DAY:
		return i.DumpByDay()
	default:
		panic("Unknown method")
	}
}
