package exifsort

import (
	"fmt"
)

type Method int

const (
	MethodYear  Method = iota // Year : dst -> year-> media
	MethodMonth               // Year : dst -> year-> month -> media
	MethodDay                 // Year : dst -> year-> month -> day -> media
	MethodNone
)

func (m Method) String() string {
	return [...]string{"year", "month", "day", "none"}[m]
}

func Methods() []Method {
	return []Method{
		MethodYear,
		MethodMonth,
		MethodDay,
	}
}

func MethodParse(str string) (Method, error) {
	for _, val := range Methods() {
		if str == val.String() {
			return val, nil
		}
	}

	return MethodNone, fmt.Errorf("invalid method %s", str)
}

type Action int

const (
	ActionCopy Action = iota // copying
	ActionMove               // moving
	ActionNone
)

func Actions() []Action {
	return []Action{
		ActionCopy,
		ActionMove,
	}
}

func (a Action) String() string {
	return [...]string{"copy", "move", "none"}[a]
}

func ActionParse(str string) (Action, error) {
	for _, val := range Actions() {
		if str == val.String() {
			return val, nil
		}
	}

	return ActionNone, fmt.Errorf("invalid action %s", str)
}
