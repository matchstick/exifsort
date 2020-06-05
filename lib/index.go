package exifsort

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"
)

type indexError struct {
	prob string
}

func (e indexError) Error() string {
	return e.prob
}

// The goal of the index system is to be able to accept as input:
// a) media pathname
// b) time it should be sorted to
// c) method to sort (by year, by month, by day)

// The index will store the original path, and the basename in a time based
// directory structure. We need top handle collisions and duplicates. Hence the
// data structure.

// MediaMap:
// key   == new name for file. It could be the same as old basename or modified
//          for collision.
// value == Holds the original full path.

type mediaMap map[string]string

// Nodes can optionally be a leaf (where it would populate it's media)
// or an intermediary where it would populate it's children.
type node struct {
	media    mediaMap
	children nodeMap
	id       int
}

// The root node has no id, so we have this sentinel value.
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
	var keys = make([]string, len(m))
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

// Add a file to the mediaMap. It needs to handle collisions, duplicates, etc.
func (n *node) mediaAdd(path string) error {
	// Use collisionRename to find a name that won't collide with others.
	base, err := uniqueName(path, func(filename string) string { return n.media[filename] })
	if err != nil {
		return err
	}

	n.media[base] = path

	return nil
}

// yearIndex will sort the paths by year.
// It's node's have no children.
type yearIndex struct {
	n node
}

func (y *yearIndex) PathStr(time time.Time, base string) string {
	year := fmt.Sprintf("%04d", time.Year())
	return filepath.Join(year, base)
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
	year := fmt.Sprintf("%04d", time.Year())                     // Year label
	month := fmt.Sprintf("%04d_%02d", time.Year(), time.Month()) // Month label

	return filepath.Join(year, month, base)
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
	year := fmt.Sprintf("%04d", time.Year())                     // Year label
	month := fmt.Sprintf("%04d_%02d", time.Year(), time.Month()) // Month label
	day := fmt.Sprintf("%04d_%02d_%02d",
		time.Year(), time.Month(), time.Day()) // Day label

	return filepath.Join(year, month, day, base)
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
	Get(string) (string, bool)
	GetAll() mediaMap
	PathStr(time.Time, string) string
	Put(string, time.Time) error
	String() string
}

func newIndex(method int) (index, error) {
	switch method {
	case MethodYear:
		var y yearIndex

		y.n.init(rootIndex)

		return &y, nil
	case MethodMonth:
		var m monthIndex

		m.n.init(rootIndex)

		return &m, nil
	case MethodDay:
		var d dayIndex

		d.n.init(rootIndex)

		return &d, nil
	default:
		errStr := fmt.Sprintf("Invalid method %d\n", method)
		return nil, &indexError{errStr}
	}
}
