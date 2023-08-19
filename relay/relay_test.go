package relay

import (
	"testing"
)

// test function
func TestMaintainConnection(t *testing.T) {
	actualInt := MaintainConnection()
	expectedInt := 1
	if actualInt != expectedInt {
		t.Errorf("Expected Int(%d) is not same as"+
			" actual string (%d)", expectedInt, actualInt)
	}
}
