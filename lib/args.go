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

func (e *parseError) argChoices() string {
	var choicesStr = make([]string, len(e.choices))

	for _, str := range e.choices {
		str = fmt.Sprintf("\"%s\"", str)
		choicesStr = append(choicesStr, str)
	}

	return strings.Join(choicesStr, ",")
}

func (e *parseError) Error() string {
	return fmt.Sprintf("must be one of [%s] (case insensitive)", e.argChoices())
}

// ParseMethod returns the constant value of the str
// Input is case insensitive.
func ParseMethod(str string) (int, error) {
	var methodMap = map[int]string{
		MethodYear:  "Year",
		MethodMonth: "Month",
		MethodDay:   "Day",
	}

	val := parseArg(str, methodMap)
	if val == parseUnknown {
		return MethodNone, &parseError{methodMap}
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

// ParseAction returns the constant value of the str
// Input is case insensitive.
func ParseAction(str string) (int, error) {
	var actionMap = map[int]string{
		ActionCopy: "Copy",
		ActionMove: "Move",
	}

	val := parseArg(str, actionMap)
	if val == parseUnknown {
		return ActionNone, &parseError{actionMap}
	}

	return val, nil
}
