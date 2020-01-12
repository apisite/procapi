package procapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/lib/pq"
)

// PGType implements marshalling between API and nonstandart postgresql types
type PGType struct{}

func (t PGType) Marshal(typ string, val interface{}) (rv interface{}, err error) {
	if strings.HasSuffix(typ, "[]") {
		rv = pq.Array(val)
	} else if typ == "money" {
		rv = fmt.Sprintf("%v", val)
	} else {
		rv = val
	}
	return
}

func (t PGType) Unmarshal(typ string, val interface{}) (rv interface{}, err error) {
	switch typ {
	case "text[]":
		src := val.(*pgtype.TextArray)
		var x []string
		err = src.AssignTo(&x)
		if err != nil {
			err = fmt.Errorf("Type %s, Val %+v error: %w", typ, val, err)
		}
		rv = &x
	case "integer[]":
		// TODO: parsed wrong
		src, ok := val.(*pgtype.Int4Array)
		if !ok {
			err = fmt.Errorf("(2)Type %s, Val %+v error: %w", typ, val, err)
		}
		var x []int
		err = src.AssignTo(&x)
		if err != nil {
			err = fmt.Errorf("Type %s, Val %+v Var: %#v error: %w", typ, val, x, err)
		}
		rv = x
	case "inet":
		// TODO: parsed wrong
		rv = val
	case "character":
		switch val.(type) {
		case string:
			rv = val.(string)
		default:
			rv = val.([]byte)
		}
	case "name":
		rv = string(val.([]byte))
	case "interval":
		src := val.(*pgtype.Interval)
		var x time.Duration
		x = time.Duration(src.Microseconds) * time.Microsecond // TODO: +Days* + Months*
		rv = x.String()
	case "money":
		// TODO: parsed wrong
		// TODO: convert "$5,678.90" -> 5678.90 ?
		rv = string(val.([]byte))
	case "numeric":
		src := val.(*pgtype.Numeric)
		var x float64
		src.AssignTo(&x)
		rv = x
	default:
		rv = val
		//fmt.Printf("==Default %s: %v\n", typ, val)
	}
	return
}
