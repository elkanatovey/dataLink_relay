package relay

import (
	"context"
	"mbg-relay/relayconn/api"
	"sync"
	"testing"
)

var exporterName = "aa"
var connReq1 = api.ConnectionRequest{
	Data:       "Some data",
	ImporterID: "123",
	ExporterID: exporterName,
}

var connReq2 = api.ConnectionRequest{
	Data:       "Some data",
	ImporterID: "abc",
	ExporterID: exporterName,
}

func TestExporterDB_NotifyExporter(t *testing.T) {
	type fields struct {
		exporters map[string]*Exporter
		mx        sync.RWMutex
	}
	type args struct {
		id  string
		msg *ImporterData
	}
	tests := []struct {
		name    string
		fields  fields
		args    []args
		wantErr bool
	}{

		{
			name:   "basic_test",
			fields: fields{map[string]*Exporter{exporterName: InitExporter(context.TODO())}, sync.RWMutex{}},
			args: []args{
				{exporterName, InitImporterData(connReq1)},
				{exporterName, InitImporterData(connReq2)},
			},
			wantErr: false,
		},
		{
			name:   "basic_test2",
			fields: fields{map[string]*Exporter{exporterName: InitExporter(context.TODO())}, sync.RWMutex{}},
			args: []args{
				{exporterName, InitImporterData(connReq2)},
			},
			wantErr: false,
		},
		{
			name:   "multiple_notifications_test",
			fields: fields{map[string]*Exporter{exporterName: InitExporter(context.TODO())}, sync.RWMutex{}},
			args: []args{
				{exporterName, InitImporterData(connReq1)},
				{exporterName, InitImporterData(connReq2)},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &ExporterDB{
				exporters: tt.fields.exporters,
				mx:        sync.RWMutex{},
			}
			for _, request := range tt.args {
				if err := db.NotifyExporter(request.id, request.msg); (err != nil) != tt.wantErr {
					t.Errorf("NotifyExporter() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
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
