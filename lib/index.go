package exifSort

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

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
	media   mediaMap
	children bucketMap
	id      int
}

func (b *bucket) Media() mediaMap {
	return b.media
}

func (b *bucket) Init(id int) {
	b.media = make(mediaMap)
	b.children = make(bucketMap)
	b.id = id
}

// Return a sorted array of keys.
func (b *bucket) ChildrenKeys() []int {
	var keys []int
	for k := range b.children {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// Returns a _sorted_ key list.
func (b *bucket) SortMediaKeys(m mediaMap) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// Get Bucket will retrieve the bucket child based on id but if none is there
// it will create one.
func (b *bucket) GetBucket(id int) bucket {
	var retBucket bucket
	retBucket, present := b.children[id]
	if present == false {
		retBucket.Init(id)
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
func (b *bucket) MediaAdd(path string) {
	var base = filepath.Base(path)
	_, present := b.media[base]
	fmt.Printf("Adding %s\n",path)
	if present {
		fmt.Printf("collision with %s\n",path)
		//TODO test for duplicate time and contents
		base = b.mediaCollisionName(base)
		fmt.Printf("new base %s\n",base)
	}
	b.media[base] = path
}

// yearIndex will sort the paths by year. Each of its bucket's have no
// children.
type yearIndex struct {
	b bucket
}

func (y *yearIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%s", time.Year(), base)
}

func (y *yearIndex) Put(path string, time time.Time) {
	yearBucket := y.b.GetBucket(time.Year())
	yearBucket.MediaAdd(path)
}

func (y *yearIndex) Get(path string) (string, bool) {
	for year, yearBucket := range y.b.children {
		for base, origPath := range yearBucket.Media() {
			if path == origPath {
				t := time.Date(year,1,1,
						1,1,1,1, time.Local)
				return y.PathStr(t, base), true
			}
		}
	}
	return "", false
}

func (y *yearIndex) GetAll() mediaMap {
	var retMap = make(mediaMap)
	for year, yearBucket := range y.b.children {
		media := yearBucket.Media()
		for base, oldPath := range media {
			time := time.Date(year,1,1,1,1,1,1, time.Local)
			path := y.PathStr(time, base)
			retMap[path] = oldPath
		}
	}
	return retMap
}

func (y yearIndex) String() string {
	var retStr string
	media := y.GetAll()
	keys := y.b.SortMediaKeys(media)
	for _, newPath := range keys {
		oldPath := media[newPath]
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
	monthBucket.MediaAdd(path)
}

func (m *monthIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%02d/%s", time.Year(), time.Month(), base)
}

func (m *monthIndex) Get(path string) (string, bool) {
	for year, yearBucket := range m.b.children {
		for month, monthBucket := range yearBucket.children {
			for base, origPath := range monthBucket.media {
				if path == origPath {
					time := time.Date(year,time.Month(month),1,1,1,1,1, time.Local)
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
			media := monthBucket.Media()
			for base, oldPath := range media {
				time := time.Date(year, time.Month(month),1,1,1,1,1, time.Local)
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
	keys := m.b.SortMediaKeys(media)
	for _, newPath := range keys {
		oldPath := media[newPath]
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
	dayBucket.MediaAdd(path)
}

func (d *dayIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%02d/%02d/%s", time.Year(), time.Month(), time.Day(), base)
}

func (d *dayIndex) Get(path string) (string, bool) {
	for year, yearBucket := range d.b.children {
		for month, monthBucket := range yearBucket.children {
			for day, dayBucket := range monthBucket.children {
				for base, origPath := range dayBucket.media {
					if path == origPath {
						time := time.Date(year, time.Month(month),day,1,1,1,1, time.Local)
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
				media := dayBucket.Media()
				for base, oldPath := range media {
					time := time.Date(year, time.Month(month),day,1,1,1,1, time.Local)
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
	keys := d.b.SortMediaKeys(media)
	for _, newPath := range keys {
		oldPath:= media[newPath]
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type index interface {
	Put(string, time.Time)
	Get(string) (string, bool)
	GetAll() mediaMap
	PathStr(time.Time, string) string
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
