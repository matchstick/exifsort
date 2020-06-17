package exifsort

import (
	"strings"
	"testing"
)

func TestParseMethod(t *testing.T) {
	var testMethods = map[string]int{
		"Year":  MethodYear,
		"year":  MethodYear,
		"YEAR":  MethodYear,
		"yEaR":  MethodYear,
		"Month": MethodMonth,
		"month": MethodMonth,
		"MONTH": MethodMonth,
		"MoNtH": MethodMonth,
		"Day":   MethodDay,
		"day":   MethodDay,
		"DAY":   MethodDay,
		"DaY":   MethodDay,
	}

	for str, val := range testMethods {
		retVal, err := MethodParse(str)
		if retVal != val {
			t.Errorf("Method %s does not match val", str)
		}

		if err != nil {
			t.Errorf("Did not expect error for %s", str)
		}
	}

	badStr := "Glabble"

	retVal, err := MethodParse(badStr)
	if !strings.Contains(err.Error(), "must be one of") {
		t.Errorf("Unexpected error string %s", err.Error())
	}

	if retVal != MethodNone {
		t.Errorf("Method %s does not match val", badStr)
	}

	if err == nil {
		t.Errorf("Did not expect error for %s", badStr)
	}
}

func TestParseAction(t *testing.T) {
	var testActions = map[string]int{
		"Copy": ActionCopy,
		"copy": ActionCopy,
		"COPY": ActionCopy,
		"cOpY": ActionCopy,
		"Move": ActionMove,
		"move": ActionMove,
		"MOVE": ActionMove,
		"MoVe": ActionMove,
	}

	for str, val := range testActions {
		retVal, err := ActionParse(str)
		if retVal != val {
			t.Errorf("Method %s does not match val", str)
		}

		if err != nil {
			t.Errorf("Did not expect error for %s", str)
		}
	}

	badStr := "Glabble"

	retVal, err := ActionParse(badStr)
	if !strings.Contains(err.Error(), "must be one of") {
		t.Errorf("Unexpected error string %s", err.Error())
	}

	if retVal != ActionNone {
		t.Errorf("Method %s does not match val", badStr)
	}

	if err == nil {
		t.Errorf("Did not expect error for %s", badStr)
	}
}
