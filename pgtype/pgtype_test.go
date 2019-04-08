package pgtype

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	want := &PGType{}
	if got := New(); !reflect.DeepEqual(got, want) {
		t.Errorf("New() = %v, want %v", got, want)
	}
}

func TestPGType_Marshal(t *testing.T) {
	type args struct {
		typ string
		v   interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantA   interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		pgt := PGType{}
		gotA, err := pgt.Marshal(tt.args.typ, tt.args.v)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. PGType.Marshal() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(gotA, tt.wantA) {
			t.Errorf("%q. PGType.Marshal() = %v, want %v", tt.name, gotA, tt.wantA)
		}
	}
}

func TestPGType_Unmarshal(t *testing.T) {
	type args struct {
		typ  string
		data interface{}
	}
	tests := []struct {
		name    string
		t       PGType
		args    args
		wantRv  interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		pgt := PGType{}
		gotRv, err := pgt.Unmarshal(tt.args.typ, tt.args.data)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. PGType.Unmarshal() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(gotRv, tt.wantRv) {
			t.Errorf("%q. PGType.Unmarshal() = %v, want %v", tt.name, gotRv, tt.wantRv)
		}
	}
}
