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

func (i *index) PutMediaByYear(path string, time time.Time) {
	yearIndex := i.GetIndex(time.Year())
	yearIndex.AddPath(path)
}

func (i *index) PutMediaByMonth(path string, time time.Time) {
	yearIndex := i.GetIndex(time.Year())
	monthIndex := yearIndex.GetIndex(int(time.Month()))
	monthIndex.AddPath(path)
}

func (i *index) PutMediaByDay(path string, time time.Time) {
	yearIndex := i.GetIndex(time.Year())
	monthIndex := yearIndex.GetIndex(int(time.Month()))
	dayIndex := monthIndex.GetIndex(time.Day())
	dayIndex.AddPath(path)
}

// I hate these switches. Will fix with next check in
func (i *index) Put(path string, time time.Time) {
	switch i.method {
	case METHOD_YEAR:
		i.PutMediaByYear(path, time)
	case METHOD_MONTH:
		i.PutMediaByMonth(path, time)
	case METHOD_DAY:
		i.PutMediaByDay(path, time)
	default:
		panic("Unknown method")
	}
}

func (i *index) NewPathByYear(year int, base string) string {
	return fmt.Sprintf("%04d/%s", year, base)
}

func (i *index) NewPathByMonth(year int, month int, base string) string {
	return fmt.Sprintf("%04d/%02d/%s", year, month, base)
}

func (i *index) NewPathByDay(year int, month int, day int, base string) string {
	return fmt.Sprintf("%04d/%02d/%02d/%s", year, month, day, base)
}

func (i *index) GetByYear(path string) (string, bool) {
	for _, year := range i.entries {
		yearIndex := i.GetIndex(year.id)
		for base, origPath := range yearIndex.Media() {
			if path == origPath {
				return i.NewPathByYear(year.id, base), true
			}
		}
	}
	return "", false
}

func (i *index) GetByMonth(path string) (string, bool) {
	for _, year := range i.entries {
		yearIndex := i.GetIndex(year.id)
		for _, month := range yearIndex.entries {
			monthIndex := yearIndex.GetIndex(month.id)
			for base, origPath := range monthIndex.media {
				if path == origPath {
					return i.NewPathByMonth(year.id, month.id, base), true
				}
			}
		}
	}
	return "", false
}

func (i *index) GetByDay(path string) (string, bool) {
	for _, year := range i.entries {
		yearIndex := i.GetIndex(year.id)
		for _, month := range yearIndex.entries {
			monthIndex := yearIndex.GetIndex(month.id)
			for _, day := range monthIndex.entries {
				dayIndex := monthIndex.GetIndex(day.id)
				for base, origPath := range dayIndex.media {
					if path == origPath {
						return i.NewPathByDay(year.id, month.id, day.id, base), true
					}
				}
			}
		}
	}
	return "", false
}

func (i *index) Get(path string) (string, bool) {
	switch i.method {
	case METHOD_YEAR:
		return i.GetByYear(path)
	case METHOD_MONTH:
		return i.GetByMonth(path)
	case METHOD_DAY:
		return i.GetByDay(path)
	default:
		panic("Unknown method")
	}
}


func (i *index) GetAllByYear() mediaMap {
	var retMap = make(mediaMap)
	for _, year := range i.EntriesKeys() {
		yearIndex := i.GetIndex(year)
		media := yearIndex.Media()
		for _, base := range media.Keys() {
			path:= i.NewPathByYear(year, base)
			retMap[path] = media[base]
		}
	}
	return retMap
}

func (i *index) GetAllByMonth() mediaMap {
	var retMap = make(mediaMap)
	for _, year := range i.EntriesKeys() {
		yearIndex := i.GetIndex(year)
		for _, month := range yearIndex.EntriesKeys() {
			monthIndex := yearIndex.GetIndex(month)
			media := monthIndex.Media()
			for _, base := range media.Keys() {
				path := i.NewPathByMonth(year, month, base)
				retMap[path] = media[base]
			}
		}
	}
	return retMap
}

func (i *index) GetAllByDay() mediaMap {
	var retMap = make(mediaMap)
	for _, year := range i.EntriesKeys() {
		yearIndex := i.GetIndex(year)
		for _, month := range yearIndex.EntriesKeys() {
			monthIndex := yearIndex.GetIndex(month)
			for _, day := range monthIndex.EntriesKeys() {
				dayIndex := monthIndex.GetIndex(month)
				media := dayIndex.Media()
				for _, base := range media.Keys() {
					path := i.NewPathByDay(year, month, day, base)
					retMap[path] = media[base]
				}
			}
		}
	}
	return retMap
}

func (i *index) GetAll() mediaMap {
	switch i.method {
	case METHOD_YEAR:
		return i.GetAllByYear()
	case METHOD_MONTH:
		return i.GetAllByMonth()
	case METHOD_DAY:
		return i.GetAllByDay()
	default:
		panic("Unknown method")
	}
}

// I hate these switches. Will fix with next check in
func (i index) String() string {
	var retMap = make(mediaMap)
	var retStr string
	switch i.method {
	case METHOD_YEAR:
		retMap = i.GetAllByYear()
	case METHOD_MONTH:
		retMap = i.GetAllByMonth()
	case METHOD_DAY:
		retMap = i.GetAllByDay()
	default:
		panic("Unknown method")
	}

	for datePath, oldPath := range retMap {
		retStr += fmt.Sprintf("%s => %s\n", oldPath, datePath)
	}
	return retStr
}
