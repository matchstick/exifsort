package exifsort

import (
	"fmt"
	"strings"
	"testing"
)

func TestSkipFileType(t *testing.T) {
	// Try just gobo.<suffix>
	for suffix := range mediaSuffixMap() {
		goodInput := fmt.Sprintf("gobo.%s", suffix)
		if skipFileType(goodInput) {
			t.Errorf("Expected True for %s\n", goodInput)
		}
	}
	// Try a simple upper case just gobo.<suffix>
	for suffix := range mediaSuffixMap() {
		goodInput := strings.ToUpper(fmt.Sprintf("gobo.%s", suffix))
		if skipFileType(goodInput) {
			t.Errorf("Expected True for %s\n", goodInput)
		}
	}

	// Try with many "." hey.gobo.<suffix>
	for suffix := range mediaSuffixMap() {
		goodInput := fmt.Sprintf("hey.gobo.%s", suffix)
		if skipFileType(goodInput) {
			t.Errorf("Expected True for %s\n", goodInput)
		}
	}

	badInput := "gobobob.."
	if skipFileType(badInput) == false {
		t.Errorf("Expected False for %s\n", badInput)
	}

	badInput = "gobo"
	if skipFileType(badInput) == false {
		t.Errorf("Expected False for %s\n", badInput)
	}

	// Try ".." at the end.<suffix>
	for suffix := range mediaSuffixMap() {
		badInput := fmt.Sprintf("gobo.%s..", suffix)
		if skipFileType(badInput) == false {
			t.Errorf("Expected False for %s\n", badInput)
		}
	}
}

func TestSkipSynologyTypes(t *testing.T) {
	badInput := "@eaDir"
	if skipFileType(badInput) == false {
		t.Errorf("Expected False for %s\n", badInput)
	}

	badInput = "@syno"
	if skipFileType(badInput) == false {
		t.Errorf("Expected False for %s\n", badInput)
	}
	badInput = "synofile_thumb"
	if skipFileType(badInput) == false {
		t.Errorf("Expected False for %s\n", badInput)
	}
}

