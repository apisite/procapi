package pgtype

import (
	"reflect"
	"testing"

	//		"github.com/jackc/pgx/pgtype"
	//	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
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
		typ     string
		val     interface{}
		rv      interface{}
		wantErr bool
	}{
		{typ: "text[]",
			val: []string{`{"b":2}`, `{"c":3}`},
			rv:  pq.Array([]string{`{"b":2}`, `{"c":3}`}),
		},
		{typ: "integer[]", val: []int{1, 2, 3}, rv: pq.Array([]int{1, 2, 3})},
		{typ: "json", val: []int{1, 2, 3}, rv: []byte("[1,2,3]")},
		{typ: "jsonb", val: []int{1, 2, 3}, rv: []byte("[1,2,3]")},
		{typ: "text", val: "test", rv: "test"},
	}
	for _, tt := range tests {
		pgt := PGType{}
		if tt.name == "" {
			tt.name = tt.typ
		}
		gotA, err := pgt.Marshal(tt.typ, tt.val)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. PGType.Marshal() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(gotA, tt.rv) {
			t.Errorf("%q. PGType.Marshal() = %v, want %v", tt.name, gotA, tt.rv)
		}
	}
}

func TestPGType_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		typ     string
		val     interface{}
		rv      interface{}
		wantErr bool
	}{
		{typ: "integer[]", val: []uint8{123, 49, 44, 50, 44, 51, 125}, rv: pq.Int64Array([]int64{1, 2, 3})},
		{typ: "text[]", val: []uint8{123, 34, 123, 92, 34, 98, 92, 34, 58, 50, 125, 34, 44, 34, 123, 92, 34, 99, 92, 34, 58, 51, 125, 34, 125}, rv: pq.StringArray([]string{`{"b":2}`, `{"c":3}`})},
		//	{typ: "json", val: []uint8{123, 34, 98, 34, 58, 50, 44, 34, 112, 114, 101, 99, 111, 109, 112, 117, 116, 101, 100, 34, 58, 116, 114, 117, 101, 125}, rv: 1},
		//	{typ: "jsonb", val: []uint8{123, 34, 98, 34, 58, 32, 50, 44, 32, 34, 112, 114, 101, 99, 111, 109, 112, 117, 116, 101, 100, 34, 58, 32, 116, 114, 117, 101, 125}, rv: 1},
		{typ: "character", val: []uint8{208, 185}, rv: "й"},
		//	{typ: "name", val: []uint8{208, 185}, rv: 1},
		{typ: "interval", val: []uint8{48, 48, 58, 48, 48, 58, 49, 48}, rv: "10s"},
		{typ: "inet", val: []uint8{49, 50, 55, 46, 48, 46, 49, 46, 50, 47, 56}, rv: "127.0.1.2/8"},
		{typ: "money", val: []uint8{36, 53, 44, 54, 55, 56, 46, 57, 48}, rv: "$5,678.90"},
		//	{typ: "numeric", val: []uint8{55, 56, 57, 48, 46, 49, 50, 51, 52}, rv: float64(7890.1234)},
	}
	for _, tt := range tests {
		pgt := PGType{}
		if tt.name == "" {
			tt.name = tt.typ
		}
		gotRv, err := pgt.Unmarshal(tt.typ, tt.val)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. PGType.Unmarshal() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(gotRv, tt.rv) {
			t.Errorf("%q. PGType.Unmarshal(%T) = %v, want %v", tt.name, gotRv, gotRv, tt.rv)
		}
	}
}
