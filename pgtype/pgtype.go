// Package pgtype implements marshalling between API and nonstandart postgresql types
package pgtype

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
)

type PGType struct{}

func New() *PGType {
	return &PGType{}
}

func (t PGType) Marshal(typ string, v interface{}) (a interface{}, err error) {

	if strings.HasSuffix(typ, "[]") {
		a = pq.Array(v)
	} else if typ == "json" || typ == "jsonb" {
		a, err = json.Marshal(v)
	} else {
		a = v
	}
	return
}

func (t PGType) Unmarshal(typ string, data interface{}) (rv interface{}, err error) {
	switch typ {
	case "text[]":
		var x pq.StringArray
		x.Scan(data)
		rv = x
	case "integer[]":
		var x pq.Int64Array
		x.Scan(data)
		rv = x
	case "json":
		var x types.JSONText
		x.Scan(data)
		rv = json.RawMessage(x)
	case "jsonb":
		var x types.JSONText
		x.Scan(data)
		rv = json.RawMessage(x)
	case "character":
		rv = string(data.([]byte))
	case "interval":
		var src pgtype.Interval
		src.DecodeText(nil, data.([]byte))
		var x time.Duration
		x = time.Duration(src.Microseconds) * time.Microsecond // TODO: +Days* + Months*
		rv = x.String()
	case "inet":
		rv = string(data.([]byte))
	case "money":
		// TODO: convert "$5,678.90" -> 5678.90 ?
		rv = string(data.([]byte))
	case "numeric":
		var src pgtype.Numeric
		src.DecodeText(nil, data.([]byte))
		var x float64
		src.AssignTo(&x)
		rv = x
	default:
		rv = data
		//					log.Printf("Skip %s: %v", c.Name, c.Type)
	}
	return rv, nil
}
