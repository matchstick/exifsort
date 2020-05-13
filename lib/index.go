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

type bucketMap map[int]bucket

const ROOT_INDEX = -1

type bucket struct {
	media   mediaMap
	entries bucketMap
	id      int
	// I hate having this switch I must fix
	method int
}

func (b *bucket) Media() mediaMap {
	return b.media
}

func (b *bucket) Init(id int) {
	b.media = make(mediaMap)
	b.entries = make(bucketMap)
	b.id = id
}

func (b *bucket) EntriesKeys() []int {
	var keys []int
	for k := range b.entries {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (b *bucket) GetBucket(id int) bucket {
	var retBucket bucket
	retBucket, present := b.entries[id]
	if present == false {
		retBucket.Init(id)
		b.entries[id] = retBucket
	}
	return retBucket
}

func (b *bucket) AddPath(path string) {
	b.media = mediaMapAdd(b.media, path)
}

type yearIndex struct {
	b bucket
}

func (y *yearIndex) NewPath(year int, base string) string {
	return fmt.Sprintf("%04d/%s", year, base)
}

func (y *yearIndex) Put(path string, time time.Time) {
	yearBucket := y.b.GetBucket(time.Year())
	yearBucket.AddPath(path)
}

func (y *yearIndex) Get(path string) (string, bool) {
	for year, yearBucket := range y.b.entries {
		for base, origPath := range yearBucket.Media() {
			if path == origPath {
				return y.NewPath(year, base), true
			}
		}
	}
	return "", false
}

func (y *yearIndex) GetAll() mediaMap {
	var retMap = make(mediaMap)
	for _, year := range y.b.EntriesKeys() {
		yearBucket := y.b.GetBucket(year)
		media := yearBucket.Media()
		for _, base := range media.Keys() {
			path := y.NewPath(year, base)
			retMap[path] = media[base]
		}
	}
	return retMap
}

func (y yearIndex) String() string {
	var retStr string
	var retMap = y.GetAll()
	for newPath, oldPath := range retMap {
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type monthIndex struct {
	b bucket
}

func (m *monthIndex) Put(path string, time time.Time) {
	yearBucket := m.b.GetBucket(time.Year())
	monthBucket := yearBucket.GetBucket(int(time.Month()))
	monthBucket.AddPath(path)
}

func (m *monthIndex) NewPath(year int, month int, base string) string {
	return fmt.Sprintf("%04d/%02d/%s", year, month, base)
}

func (m *monthIndex) Get(path string) (string, bool) {
	for year, yearBucket := range m.b.entries {
		for month, monthBucket := range yearBucket.entries {
			for base, origPath := range monthBucket.media {
				if path == origPath {
					return m.NewPath(year, month, base), true
				}
			}
		}
	}
	return "", false
}

func (m *monthIndex) GetAll() mediaMap {
	var retMap = make(mediaMap)
	for _, year := range m.b.EntriesKeys() {
		yearBucket := m.b.GetBucket(year)
		for _, month := range yearBucket.EntriesKeys() {
			monthBucket := yearBucket.GetBucket(month)
			media := monthBucket.Media()
			for _, base := range media.Keys() {
				path := m.NewPath(year, month, base)
				retMap[path] = media[base]
			}
		}
	}
	return retMap
}

func (m monthIndex) String() string {
	var retStr string
	var retMap = m.GetAll()
	for newPath, oldPath := range retMap {
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type dayIndex struct {
	b bucket
}

func (d *dayIndex) Put(path string, time time.Time) {
	yearBucket := d.b.GetBucket(time.Year())
	monthBucket := yearBucket.GetBucket(int(time.Month()))
	dayBucket := monthBucket.GetBucket(time.Day())
	dayBucket.AddPath(path)
}

func (d *dayIndex) NewPath(year int, month int, day int, base string) string {
	return fmt.Sprintf("%04d/%02d/%02d/%s", year, month, day, base)
}

func (d *dayIndex) Get(path string) (string, bool) {
	for year, yearBucket := range d.b.entries {
		for month, monthBucket := range yearBucket.entries {
			for day, dayBucket := range monthBucket.entries {
				for base, origPath := range dayBucket.media {
					if path == origPath {
						return d.NewPath(year, month, day, base), true
					}
				}
			}
		}
	}
	return "", false
}

func (d *dayIndex) GetAll() mediaMap {
	var retMap = make(mediaMap)
	for _, year := range d.b.EntriesKeys() {
		yearBucket := d.b.GetBucket(year)
		for _, month := range yearBucket.EntriesKeys() {
			monthBucket := yearBucket.GetBucket(month)
			for _, day := range monthBucket.EntriesKeys() {
				dayBucket := monthBucket.GetBucket(month)
				media := dayBucket.Media()
				for _, base := range media.Keys() {
					path := d.NewPath(year, month, day, base)
					retMap[path] = media[base]
				}
			}
		}
	}
	return retMap
}

// I hate these switches. Will fix with next check in
func (d dayIndex) String() string {
	var retStr string
	var retMap = d.GetAll()
	for newPath, oldPath := range retMap {
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type index interface {
	Put(string, time.Time)
	Get(string) (string, bool)
	GetAll() mediaMap
}

func CreateIndex(method int) index {
	switch method {
	case METHOD_YEAR:
		var y yearIndex
		y.b.Init(ROOT_INDEX)
		return &y
	case METHOD_MONTH:
		var m monthIndex
		m.b.Init(ROOT_INDEX)
		return &m
	case METHOD_DAY:
		var d dayIndex
		d.b.Init(ROOT_INDEX)
		return &d
	default:
		panic("Unknown method")
	}
}
