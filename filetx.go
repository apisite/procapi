package procapi

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	//"github.com/olekukonko/tablewriter"
)

const (
	SQLInit  = `CREATE SCHEMA %s; SET SEARCH_PATH = %s, public;`
	SQLReset = `RESET SEARCH_PATH;` // TODO: reset after load
)

func loadPath(tx Tx, schema string, aliases map[string]string) error {

	alias := aliases[schema]
	sql := fmt.Sprintf(SQLInit, alias, alias)
	_, err := tx.Exec(sql)
	if err != nil {
		return err
	}

	path := filepath.Join("testdata", schema)
	files, _ := filepath.Glob(path + "/[1-7]?_*.sql") // load only files with sources
	for _, f := range files {
		if strings.Contains(f, "/1") && !(strings.Contains(f, "/18") || strings.Contains(f, "/19")) {
			// skip pkg setup files
			continue
		}
		fmt.Printf("Load %s\n", f)
		s, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		sql := string(s)
		for k, v := range aliases {
			sql = strings.ReplaceAll(sql, k+".", v+".")
		}
		_, err = tx.Exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

/*

package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
)

func main() {
    file, err := os.Open("/path/to/file.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
}
*/
/*
	table := tablewriter.NewWriter(os.Stdout)
	//able.SetHeader([]string{"Date", "Description", "CV2", "Amount"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoWrapText(false)
	table.AppendBulk(rv) // Add Bulk Data
	table.Render()
	//
*/
