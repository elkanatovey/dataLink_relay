package relay

import (
	"context"
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

func TestExporterDB(t *testing.T) {
	a := InitExporterDB()
	exp := "aa"
	a.AddExporter(exp, InitExporter(context.TODO()))
	a.RemoveExporter(exp)

	actualSize := len(a.exporters)
	expectedSize := 0
	if actualSize != expectedSize {
		t.Errorf("Expected Size(%d) is not same as"+
			" actual Size (%d)", expectedSize, actualSize)
	}
}
