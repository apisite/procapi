// Copyright 2018 Aleksei Kovrizhkin <lekovr+apisite@gmail.com>. All rights reserved.

package pgcall

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/apisite/pgcall/pgiface"
)

// CallMapAny calls postgresql stored function without metadata usage
func (srv *Server) CallMapAny(
	method string,
	args ...interface{},
) ([]map[string]interface{}, error) {

	inAssigns := make([]string, len(args))
	for i := range args {
		inAssigns[i] = fmt.Sprintf("$%d", i+1)
	}
	sql := fmt.Sprintf("select * from %s(%s)",
		method,
		strings.Join(inAssigns, ", "),
	)
	rows, err := srv.dbh.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := Maps(rows)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Call postgresql stored function
func (srv *Server) Call(r *http.Request,
	method string,
	args map[string]interface{},
) (interface{}, error) {
	// Lookup method.
	methodSpec, ok := (*srv.Methods())[method]
	if !ok {
		return nil, (&CallError{code: NotFound}).addContext("name", method)
	}

	var inAssigns []string
	var missedArgs []string
	var inVars []interface{}

	if methodSpec.In != nil {
		srv.log.Debugf("IN args: %+v", *methodSpec.In)
		for k, v := range *methodSpec.In {
			a, ok := args[k]
			if !ok {
				if v.Required {
					missedArgs = append(missedArgs, k)
				} else {
					srv.log.Debugf("Skip missed value of %s", k)
				}
				continue
			}
			if reflect.ValueOf(a).Kind() == reflect.Ptr {
				if reflect.ValueOf(a).IsNil() {
					if v.Required {
						missedArgs = append(missedArgs, k)
					} else {
						srv.log.Debugf("Skip missed ref of %s", k)
					}
					continue
				}
				a = reflect.ValueOf(a).Elem().Interface() // dereference ptr
			}
			inAssigns = append(inAssigns, fmt.Sprintf("%s %s $%d", v.Name, srv.config.ArgSyntax, len(inAssigns)+1))
			inVars = append(inVars, a)
			srv.log.Debugf("Use: %s (%+v)", k, a)
		}
	}
	if len(missedArgs) > 0 {
		return nil, (&CallError{code: ArgsMissed}).addContext("args", missedArgs)
	}

	if methodSpec.Out == nil && methodSpec.Result == nil {
		// no data returned
		sql := fmt.Sprintf("SELECT %s.%s(%s)",
			methodSpec.Class,
			methodSpec.Func,
			strings.Join(inAssigns, ", "),
		)
		ct, err := srv.dbh.Exec(sql, inVars...)
		ctra, _ := ct.RowsAffected()
		srv.log.Debugf("Rows affected: %d", ctra) // TODO: Header.Add ?
		return nil, err
	}

	var outCols []string
	if methodSpec.Out != nil {
		for _, v := range *methodSpec.Out {
			outCols = append(outCols, v.Name)
		}
	}

	from := ""
	if len(outCols) > 0 {
		from = " from "
	}

	sql := fmt.Sprintf("select %s%s%s.%s(%s)",
		strings.Join(outCols, ", "),
		from,
		methodSpec.Class,
		methodSpec.Func,
		strings.Join(inAssigns, ", "),
	)
	srv.log.Debugf("sql: %s, args: %v\n", sql, inVars)
	rows, err := srv.dbh.Query(sql, inVars...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rv interface{}

	if methodSpec.IsStruct {
		rv, err = Maps(rows)
	} else {
		rv, err = Slice(rows)
	}
	if err != nil {
		return nil, err
	}
	if !methodSpec.IsSet {
		rv1 := rv.([]interface{})
		if len(rv1) != 1 {
			return nil, errors.New("single row must be returned")
		}
		return &rv1[0], nil
	}
	return rv, nil
}

// Maps fetches []map[string]interface{} from query result
func Maps(r pgiface.Rows) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}
	fields, _ := r.Columns()
	for r.Next() {
		row, err := r.Values()
		if err != nil {
			return nil, err
		}
		rowMap := map[string]interface{}{}
		for k, v := range row {
			rowMap[fields[k]] = v
		}
		result = append(result, rowMap)
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	return result, nil
}

// Slice fetches []interface{} from query result
func Slice(r pgiface.Rows) ([]interface{}, error) {
	result := []interface{}{}
	fields, _ := r.Columns()
	if len(fields) != 1 {
		return nil, errors.New("single column must be returned")
	}
	for r.Next() {
		row, err := r.Values()
		if err != nil {
			return nil, err
		}
		result = append(result, row[0])
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	return result, nil
}
