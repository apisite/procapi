package procapi

import (
	"github.com/jackc/pgx/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

func ScanMap(rows pgx.Rows, dest *map[string]interface{}) error {
	columns := FieldNames(rows)
	/*
		values := make([]interface{}, len(columns))
		for i := range values {
			x := new(interface{})
			values[i] = x
			//values[i] = "" //interface{}{}
		}

		err := rows.Scan(values...)
	*/
	values, err := rows.Values()
	if err != nil {
		return errors.Wrap(err, "ScanMap")
	}

	for i, column := range columns {
		(*dest)[string(column)] = values[i]
		//		(*dest)[string(column)] = *(values[i].(*interface{}))
	}
	return rows.Err()

}

func ScanStruct(rows pgx.Rows, dest interface{}) error {
	m := map[string]interface{}{}
	err := ScanMap(rows, &m)
	if err != nil {
		return errors.Wrap(err, "Map")
	}
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   dest,
		TagName:  "db",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return errors.Wrap(err, "Decoder")
	}
	return decoder.Decode(m)
}

func FieldNames(rows pgx.Rows) []string {
	columns := rows.FieldDescriptions()
	names := make([]string, len(columns))
	for i, column := range columns {
		names[i] = string(column.Name)
	}
	return names
}
