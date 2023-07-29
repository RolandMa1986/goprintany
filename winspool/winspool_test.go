package winspool

import (
	"io/ioutil"
	"testing"
)

func TestRawDataToPrinter(t *testing.T) {

	type args struct {
		prn     string
		docName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Zebra Raw ZPL",
			args: args{
				prn:     "ZDesigner ZT211-300dpi ZPL",
				docName: "zebra.txt",
			},
		},
		{
			name: "Canon Raw PDF",
			args: args{
				prn:     "Canon MF110/910 Series UFRII LT",
				docName: "pdfdoc.pdf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := ioutil.ReadFile(tt.args.docName)
			if err != nil {
				t.Errorf("RawDataToPrinter() error = %v", err)
				return
			}
			if _, err := RawDataToPrinter(tt.args.prn, tt.args.docName, buf); (err != nil) != tt.wantErr {
				t.Errorf("RawDataToPrinter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
