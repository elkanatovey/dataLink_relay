package relay

import (
	"context"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/api"
	"sync"
	"testing"
)

var exporterName = "aa"
var connReq1 = api.ConnectionRequest{
	Data:     "Some data",
	ClientID: "123",
	ServerID: exporterName,
}

var connReq2 = api.ConnectionRequest{
	Data:     "Some data",
	ClientID: "abc",
	ServerID: exporterName,
}

func TestExporterDB_NotifyExporter(t *testing.T) {
	type fields struct {
		exporters map[string]*ListeningServer
		mx        sync.RWMutex
	}
	type args struct {
		id  string
		msg *ClientData
	}
	tests := []struct {
		name    string
		fields  fields
		args    []args
		wantErr bool
	}{

		{
			name:   "basic_test",
			fields: fields{map[string]*ListeningServer{exporterName: InitListeningServer(context.TODO())}, sync.RWMutex{}},
			args: []args{
				{exporterName, InitClientData(connReq1)},
				{exporterName, InitClientData(connReq2)},
			},
			wantErr: false,
		},
		{
			name:   "basic_test2",
			fields: fields{map[string]*ListeningServer{exporterName: InitListeningServer(context.TODO())}, sync.RWMutex{}},
			args: []args{
				{exporterName, InitClientData(connReq2)},
			},
			wantErr: false,
		},
		{
			name:   "multiple_notifications_test",
			fields: fields{map[string]*ListeningServer{exporterName: InitListeningServer(context.TODO())}, sync.RWMutex{}},
			args: []args{
				{exporterName, InitClientData(connReq1)},
				{exporterName, InitClientData(connReq2)},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &listeningServerDB{
				listeningServers: tt.fields.exporters,
				mx:               sync.RWMutex{},
			}
			for _, request := range tt.args {
				if err := db.NotifyListeningServer(request.id, request.msg); (err != nil) != tt.wantErr {
					t.Errorf("NotifyListeningServer() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestExporterDB(t *testing.T) {
	a := initListeningServerDB()
	exp := "aa"
	a.AddListeningServer(exp, InitListeningServer(context.TODO()))
	a.RemoveListeningServer(exp)

	actualSize := len(a.listeningServers)
	expectedSize := 0
	if actualSize != expectedSize {
		t.Errorf("Expected Size(%d) is not same as"+
			" actual Size (%d)", expectedSize, actualSize)
	}
}
