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

// The index will store the original path, and the basename in a time based
// directory structure. We need top handle collisions and duplicates. Hence the
// data structure.

// key   == new name for file. It could be the same as old basename or modified
//          for collision.
// value == original full path
type mediaMap map[string]string

// Nodes can optionally be a leaf (where it would populate it's media)
// or an intermediary where it would populate it's children.
type node struct {
	media    mediaMap
	children nodeMap
	id       int
}

// The root node has no id, so we have this sentinel value
const rootIndex = -1

type nodeMap map[int]node

func (n *node) init(id int) {
	n.media = make(mediaMap)
	n.children = make(nodeMap)
	n.id = id
}

// Returns a _sorted_ key list. No technical reason to make it a receive
// pointer.
func (n *node) sortMediaKeys(m mediaMap) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// Get Node will retrieve the node child based on id but if none is there
// it will create one.
func (n *node) getNode(id int) node {
	var retNode node
	retNode, present := n.children[id]
	if !present {
		retNode.init(id)
		n.children[id] = retNode
	}
	return retNode
}

// When you find a collision you add a counter to the name.
// So <name>.jpg => <name>_#.jpg the number increments as it may have
// multiple collisions.
func (n *node) mediaCollisionName(base string) string {
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
		_, present := n.media[newName]
		if !present {
			break
		}
	}
	return newName
}

// Add a file to the mediaMap. It needs to handle collisions, duplicates, etc.
func (n *node) mediaAdd(path string) error {
	var base = filepath.Base(path)
	storedPath, present := n.media[base]

	// Common case, no duplicates or collisions.
	if !present {
		n.media[base] = path
		return nil
	}

	// Check for same contents
	cmp := equalfile.New(nil, equalfile.Options{})
	equal, err := cmp.CompareFile(path, storedPath)
	if err != nil {
		return err
	}

	if equal {
		return fmt.Errorf("%s is a duplicate of the already stored media %s", path, storedPath)
	}

	// If it has the same name as is not the same file we should add it
	// with a new base name to not collide.
	base = n.mediaCollisionName(base)
	n.media[base] = path
	return nil
}

// yearIndex will sort the paths by year. 
// It's node's have no children.
type yearIndex struct {
	n node
}

func (y *yearIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%s", time.Year(), base)
}

func (y *yearIndex) Put(path string, time time.Time) error {
	yearNode := y.n.getNode(time.Year())
	return yearNode.mediaAdd(path)
}

func (y *yearIndex) Get(path string) (string, bool) {
	soughtBase := filepath.Base(path)
	for year, yearNode := range y.n.children {
		for base := range yearNode.media {
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
	for year, yearNode := range y.n.children {
		media := yearNode.media
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
	keys := y.n.sortMediaKeys(media)
	for _, newPath := range keys {
		oldPath := media[newPath]
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type monthIndex struct {
	n node
}

func (m *monthIndex) Put(path string, time time.Time) error {
	yearNode := m.n.getNode(time.Year())
	monthNode := yearNode.getNode(int(time.Month()))
	return monthNode.mediaAdd(path)
}

func (m *monthIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%04d_%02d/%s",
			time.Year(), // Year label
			time.Year(), time.Month(), // Month label
			base)
}

func (m *monthIndex) Get(path string) (string, bool) {
	soughtBase := filepath.Base(path)
	for year, yearNode := range m.n.children {
		for month, monthNode := range yearNode.children {
			for base := range monthNode.media {
				if base == soughtBase {
					time := time.Date(year,
							time.Month(month),
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
	for year, yearNode := range m.n.children {
		for month, monthNode := range yearNode.children {
			media := monthNode.media
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
	keys := m.n.sortMediaKeys(media)
	for _, newPath := range keys {
		oldPath := media[newPath]
		retStr += fmt.Sprintf("%s => %s\n", oldPath, newPath)
	}
	return retStr
}

type dayIndex struct {
	n node
}

func (d *dayIndex) Put(path string, time time.Time) error {
	yearNode := d.n.getNode(time.Year())
	monthNode := yearNode.getNode(int(time.Month()))
	dayNode := monthNode.getNode(time.Day())
	return dayNode.mediaAdd(path)
}

func (d *dayIndex) PathStr(time time.Time, base string) string {
	return fmt.Sprintf("%04d/%04d_%02d/%04d_%02d_%02d/%s",
				time.Year(),
				time.Year(), time.Month(),
				time.Year(), time.Month(), time.Day(),
				base)
}

func (d *dayIndex) Get(path string) (string, bool) {
	soughtBase := filepath.Base(path)
	for year, yearNode := range d.n.children {
		for month, monthNode := range yearNode.children {
			for day, dayNode := range monthNode.children {
				for base := range dayNode.media {
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
	for year, yearNode := range d.n.children {
		for month, monthNode := range yearNode.children {
			for day, dayNode := range monthNode.children {
				media := dayNode.media
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
	keys := d.n.sortMediaKeys(media)
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

func createIndex(method int) index {
	switch method {
	case METHOD_YEAR:
		var y yearIndex
		y.n.init(rootIndex)
		return &y
	case METHOD_MONTH:
		var m monthIndex
		m.n.init(rootIndex)
		return &m
	case METHOD_DAY:
		var d dayIndex
		d.n.init(rootIndex)
		return &d
	default:
		panic("Unknown method")
	}
}
