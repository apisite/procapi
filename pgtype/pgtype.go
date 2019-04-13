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

func (t PGType) Marshal(typ string, val interface{}) (rv interface{}, err error) {
	if strings.HasSuffix(typ, "[]") {
		rv = pq.Array(val)
	} else if typ == "json" || typ == "jsonb" {
		rv, err = json.Marshal(val)
	} else {
		rv = val
	}
	return
}

func (t PGType) Unmarshal(typ string, val interface{}) (rv interface{}, err error) {
	switch typ {
	case "text[]":
		var x pq.StringArray
		x.Scan(val)
		y := make([]*string, len(x))
		for i, v := range x {
			vv := v
			y[i] = &vv
		}
		rv = y
	case "integer[]":
		var x pq.Int64Array
		x.Scan(val)
		rv = x
	case "json":
		var x types.JSONText
		x.Scan(val)
		rv = json.RawMessage(x)
	case "jsonb":
		var x types.JSONText
		x.Scan(val)
		rv = json.RawMessage(x)
	case "character":
		rv = string(val.([]byte))
	case "name":
		rv = string(val.([]byte))
	case "interval":
		var src pgtype.Interval
		src.DecodeText(nil, val.([]byte))
		var x time.Duration
		x = time.Duration(src.Microseconds) * time.Microsecond // TODO: +Days* + Months*
		rv = x.String()
	case "inet":
		rv = string(val.([]byte))
	case "money":
		// TODO: convert "$5,678.90" -> 5678.90 ?
		rv = string(val.([]byte))
	case "numeric":
		var src pgtype.Numeric
		src.DecodeText(nil, val.([]byte))
		var x float64
		src.AssignTo(&x)
		rv = x
	default:
		rv = val
		//					log.Printf("Skip %s: %v", c.Name, c.Type)
	}
	return rv, nil
}
