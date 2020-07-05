package exifsort

import (
	"fmt"
)

// Method user specifies indexing and structuring sorted directories.
type Method int

const (
	// MethodYear : dst -> year-> media
	MethodYear Method = iota
	// MethodMonth : dst -> year-> month -> media
	MethodMonth
	// MethodDay : dst -> year-> month -> day -> media
	MethodDay
	// MethodNone : Error Value
	MethodNone
)

// Returns name of method value (all lower case).
func (m Method) String() string {
	return [...]string{"year", "month", "day", "none"}[m]
}

// Methods returns all method values used excluding MethodNone.
func Methods() []Method {
	return []Method{
		MethodYear,
		MethodMonth,
		MethodDay,
	}
}

// MethodParser returns Method from string (must be lower case). Returns
// MethodNone if invalid.
func MethodParse(str string) (Method, error) {
	for _, val := range Methods() {
		if str == val.String() {
			return val, nil
		}
	}

	return MethodNone, fmt.Errorf("invalid method %s", str)
}

// Action user specifies to transfer files from src to dst.
type Action int

// User can specify how to transfer files from one directory to another.
const (
	// ActionCopy : Copying files from src to dst
	ActionCopy Action = iota
	// ActionMove : Moving files from src to dst
	ActionMove
	// ActionNone : Error Value
	ActionNone
)

// Actions returns all actions values used excluding ActionNone.
func Actions() []Action {
	return []Action{
		ActionCopy,
		ActionMove,
	}
}

// Returns name of action value (all lower case).
func (a Action) String() string {
	return [...]string{"copy", "move", "none"}[a]
}

// ActionParse returns Action from string (must be lower case). Returns
// ActionNone if invalid.
func ActionParse(str string) (Action, error) {
	for _, val := range Actions() {
		if str == val.String() {
			return val, nil
		}
	}

	return ActionNone, fmt.Errorf("invalid action %s", str)
}
