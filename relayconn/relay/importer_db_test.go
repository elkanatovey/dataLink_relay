package relay

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

type connTester struct {
	deadline time.Time
}

func (c *connTester) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (c *connTester) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (c *connTester) Close() error {

	return nil
}

func (c *connTester) LocalAddr() net.Addr {
	return nil
}

func (c *connTester) RemoteAddr() net.Addr {
	return nil
}

func (c *connTester) SetDeadline(t time.Time) error {
	return nil
}

func (c *connTester) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *connTester) SetWriteDeadline(t time.Time) error {
	return nil
}

var importerName = "bb"

func TestImporterDB_NotifyImporter(t *testing.T) {
	type fields struct {
		importers map[string]*Importer
		mx        sync.RWMutex
	}
	type args struct {
		id         string
		connection *ServerConn
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name: "basic_test",
			fields: fields{
				map[string]*Importer{importerName: InitImporter(context.TODO())},
				sync.RWMutex{},
			},
			args: args{
				importerName,
				&ServerConn{
					&connTester{},
					nil},
			},
			wantErr: false,
		},
		{
			name: "basic_test2",
			fields: fields{
				map[string]*Importer{importerName: InitImporter(context.TODO())},
				sync.RWMutex{},
			},
			args: args{
				importerName + "not",
				&ServerConn{
					&connTester{},
					nil},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &ImporterDB{
				importers: tt.fields.importers,
				mx:        sync.RWMutex{},
			}

			if err := db.NotifyImporter(tt.args.id, tt.args.connection); (err != nil) != tt.wantErr {
				t.Errorf("NotifyImpporter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
