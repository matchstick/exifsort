package exifsort

import (
	"strings"
	"testing"
)

func TestParseMethod(t *testing.T) {
	t.Parallel()
	var testMethods = map[string]Method{
		"year":  MethodYear,
		"month": MethodMonth,
		"day":   MethodDay,
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

	// Case sensitive
	for str := range testMethods {
		str = strings.ToUpper(str)

		retVal, err := MethodParse(str)
		if retVal != MethodNone {
			t.Errorf("We should have an error")
		}

		if err == nil {
			t.Errorf("Did not expect success for %s", str)
		}
	}

	badStr := "Glabble"

	retVal, err := MethodParse(badStr)
	if retVal != MethodNone {
		t.Errorf("Method %s should be MethodNone not %s", badStr, retVal)
	}

	if err == nil {
		t.Errorf("Did not expect success for %s", badStr)
	}
}

func TestParseAction(t *testing.T) {
	t.Parallel()
	var testActions = map[string]Action{
		"copy": ActionCopy,
		"move": ActionMove,
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

	// Case sensitive
	for str := range testActions {
		str = strings.ToUpper(str)

		retVal, err := ActionParse(str)
		if retVal != ActionNone {
			t.Errorf("We should have an error")
		}

		if err == nil {
			t.Errorf("Did not expect success for %s", str)
		}
	}

	badStr := "Glabble"

	retVal, err := ActionParse(badStr)
	if retVal != ActionNone {
		t.Errorf("Method %s should be MethodNone not %s", badStr, retVal)
	}

	if err == nil {
		t.Errorf("Did not expect success for %s", badStr)
	}
}
