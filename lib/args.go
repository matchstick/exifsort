package exifsort

import (
	"fmt"
	"strings"
)

// Methods to sort media files in nested directory structure.
const (
	MethodYear  = iota // Year : dst -> year-> media
	MethodMonth        // Year : dst -> year-> month -> media
	MethodDay          // Year : dst -> year-> month -> day -> media
	MethodNone
)

func MethodMap() map[int]string {
	return map[int]string{
		MethodYear:  "year",
		MethodMonth: "month",
		MethodDay:   "day",
	}
}

type argsMap map[int]string

const parseUnknown = -1

func parseArg(argStr string, argsMap map[int]string) int {
	/// lower capitalization for safe comparing
	argStr = strings.ToLower(argStr)

	for val, str := range argsMap {
		str = strings.ToLower(str)
		if argStr == str {
			return val
		}
	}

	return parseUnknown
}

type parseError struct {
	choices argsMap
}

func (e *parseError) Error() string {
	var choicesList = make([]string, len(e.choices))

	for _, str := range e.choices {
		str = fmt.Sprintf("\"%s\"", str)
		choicesList = append(choicesList, str)
	}

	choiceStr := strings.Join(choicesList, ",")

	return fmt.Sprintf("must be one of [%s] (case insensitive)", choiceStr)
}

// ParseMethod returns the constant value of the str.
// Input is case insensitive.
func MethodParse(str string) (int, error) {
	val := parseArg(str, MethodMap())
	if val == parseUnknown {
		return MethodNone, &parseError{MethodMap()}
	}

	return val, nil
}

// When sorting media you specify action to transfer files form the src to dst
// directories.
const (
	ActionCopy = iota // copying
	ActionMove        // moving
	ActionNone
)

func ActionMap() map[int]string {
	return map[int]string{
		ActionCopy: "copy",
		ActionMove: "move",
	}
}

// ParseAction returns the constant value of the str.
// Input is case insensitive.
func ActionParse(str string) (int, error) {
	val := parseArg(str, ActionMap())
	if val == parseUnknown {
		return ActionNone, &parseError{ActionMap()}
	}

	return val, nil
}
