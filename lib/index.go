package exifSort

import (
	"fmt"
	"github.com/udhos/equalfile"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// The goal of the index system is to be able to accept as input:
// a) media pathname
// b) time it should be sorted to
// c) method to sort (by year, by month, by day)

// The index will store the original path, and the basename in a time based directory structure.
// We need top handle collisions and duplicates. Hence the data structure.

// key   == new name for file. It could be the same as old basename or modified
//          for collision.
// value == original full path
type mediaMap map[string]string

// A bucket is the common "node" of the data structure.
// It can optionally be a leaf (where it would populate it's media)
// Or an intermediary where it would populate it's children.
type bucketMap map[int]bucket

const ROOT_INDEX = -1

type bucket struct {
	media    mediaMap
	children bucketMap
	id       int
}

func (b *bucket) init(id int) {
	b.media = make(mediaMap)
	b.children = make(bucketMap)
	b.id = id
}

// Return a sorted array of keys.
func (b *bucket) childrenKeys() []int {
	var keys []int
	for k := range b.children {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// Returns a _sorted_ key list. No technical reason to make it a receive pointer.
func (b *bucket) sortMediaKeys(m mediaMap) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// Get Bucket will retrieve the bucket child based on id but if none is there
// it will create one.
func (b *bucket) getBucket(id int) bucket {
	var retBucket bucket
	retBucket, present := b.children[id]
	if present == false {
		retBucket.init(id)
		b.children[id] = retBucket
	}
	return retBucket
}

// When you find a collision you add a counter to the name.
// So <name>.jpg => <name>_#.jpg the number increments as it may have
// multiple collisions.
func (b *bucket) mediaCollisionName(base string) string {
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
		_, present := b.media[newName]
		if present == false {
			break
		}
	}
	return newName
}

// Add a file to the mediaMap. It needs to handle collisions, duplicates, etc.
func (b *bucket) mediaAdd(path string) error {
	var base = filepath.Base(path)
	storedPath, present := b.media[base]

	// Common case, no duplicates or collisions.
	if present == false {
		b.media[base] = path
		return nil
	}

	// Check for same contents
	cmp := equalfile.New(nil, equalfile.Options{}) // compare using single mode
	equal, err := cmp.CompareFile(path, storedPath)
	if err != nil {
		return err
	}

	if equal {
		return fmt.Errorf("%s is a duplicate of the already stored media %s",
			path, storedPath)
	}

	// If it has the same name as is not the same file we should add it
	// with a new base name to not collide.
	base = b.mediaCollisionName(base)
	b.media[base] = path
	return nil
}

// yearIndex will sort the paths by year. Each of its bucket's have no
// children.
type yearIndex struct {
	b bucket
}

func (y *yearIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%s", time.Year(), base)
}

func (y *yearIndex) Put(path string, time time.Time) error {
	yearBucket := y.b.getBucket(time.Year())
	return yearBucket.mediaAdd(path)
}

func (y *yearIndex) Get(path string) (string, bool) {
	soughtBase := filepath.Base(path)
	for year, yearBucket := range y.b.children {
		for base, _ := range yearBucket.media {
			if base == soughtBase {
				time := time.Date(year, 1, 1,
					1, 1, 1, 1, time.Local)
				return y.PathStr(time, base), true
			}
		}
	}
	return "", false
}

func (y *yearIndex) GetAll() mediaMap {
	var retMap = make(mediaMap)
	for year, yearBucket := range y.b.children {
		media := yearBucket.media
		for base, oldPath := range media {
			time := time.Date(year, 1, 1, 1, 1, 1, 1, time.Local)
			path := y.PathStr(time, base)
			retMap[path] = oldPath
		}
	}
	return retMap
}

func (y yearIndex) String() string {
	var retStr string
	media := y.GetAll()
	keys := y.b.sortMediaKeys(media)
	for _, newPath := range keys {
		oldPath := media[newPath]
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type monthIndex struct {
	b bucket
}

func (m *monthIndex) Put(path string, time time.Time) error {
	yearBucket := m.b.getBucket(time.Year())
	monthBucket := yearBucket.getBucket(int(time.Month()))
	return monthBucket.mediaAdd(path)
}

func (m *monthIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%02d/%s", time.Year(), time.Month(), base)
}

func (m *monthIndex) Get(path string) (string, bool) {
	soughtBase := filepath.Base(path)
	for year, yearBucket := range m.b.children {
		for month, monthBucket := range yearBucket.children {
			for base, _ := range monthBucket.media {
				if base == soughtBase {
					time := time.Date(year, time.Month(month),
						1, 1, 1, 1, 1, time.Local)
					return m.PathStr(time, base), true
				}
			}
		}
	}
	return "", false
}

func (m *monthIndex) GetAll() mediaMap {
	var retMap = make(mediaMap)
	for year, yearBucket := range m.b.children {
		for month, monthBucket := range yearBucket.children {
			media := monthBucket.media
			for base, oldPath := range media {
				time := time.Date(year, time.Month(month), 1, 1, 1, 1, 1, time.Local)
				path := m.PathStr(time, base)
				retMap[path] = oldPath
			}
		}
	}
	return retMap
}

func (m monthIndex) String() string {
	var retStr string
	media := m.GetAll()
	keys := m.b.sortMediaKeys(media)
	for _, newPath := range keys {
		oldPath := media[newPath]
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type dayIndex struct {
	b bucket
}

func (d *dayIndex) Put(path string, time time.Time) error {
	yearBucket := d.b.getBucket(time.Year())
	monthBucket := yearBucket.getBucket(int(time.Month()))
	dayBucket := monthBucket.getBucket(time.Day())
	return dayBucket.mediaAdd(path)
}

func (d *dayIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%02d/%02d/%s", time.Year(), time.Month(), time.Day(), base)
}

func (d *dayIndex) Get(path string) (string, bool) {
	soughtBase := filepath.Base(path)
	for year, yearBucket := range d.b.children {
		for month, monthBucket := range yearBucket.children {
			for day, dayBucket := range monthBucket.children {
				for base, _ := range dayBucket.media {
					if base == soughtBase {
						time := time.Date(year,
							time.Month(month), day,
							1, 1, 1, 1,
							time.Local)
						return d.PathStr(time, base), true
					}
				}
			}
		}
	}
	return "", false
}

func (d *dayIndex) GetAll() mediaMap {
	var retMap = make(mediaMap)
	for year, yearBucket := range d.b.children {
		for month, monthBucket := range yearBucket.children {
			for day, dayBucket := range monthBucket.children {
				media := dayBucket.media
				for base, oldPath := range media {
					time := time.Date(year, time.Month(month), day, 1, 1, 1, 1, time.Local)
					newPath := d.PathStr(time, base)
					retMap[newPath] = oldPath
				}
			}
		}
	}
	return retMap
}

func (d dayIndex) String() string {
	var retStr string
	media := d.GetAll()
	keys := d.b.sortMediaKeys(media)
	for _, newPath := range keys {
		oldPath := media[newPath]
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type index interface {
	Put(string, time.Time) error
	Get(string) (string, bool)
	GetAll() mediaMap
	PathStr(time.Time, string) string
}

func CreateIndex(method int) index {
	switch method {
	case METHOD_YEAR:
		var y yearIndex
		y.b.init(ROOT_INDEX)
		return &y
	case METHOD_MONTH:
		var m monthIndex
		m.b.init(ROOT_INDEX)
		return &m
	case METHOD_DAY:
		var d dayIndex
		d.b.init(ROOT_INDEX)
		return &d
	default:
		panic("Unknown method")
	}
}
