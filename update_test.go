package pgcall

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	//	"path/filepath"
)

const TEST_UPDATE_ENV = "TEST_UPDATE"
const TEST_UPDATE_DIR = "testdata-new"

func checkTestUpdate(file string, data interface{}) {

	if os.Getenv(TEST_UPDATE_ENV) == "" {
		return
	}

	if _, err := os.Stat(TEST_UPDATE_DIR); os.IsNotExist(err) {
		os.Mkdir(TEST_UPDATE_DIR, os.ModePerm)
	}
	p, err := ioutil.TempFile(TEST_UPDATE_DIR, file+".")
	check(err)
	fmt.Printf("*** Writing %s\n", p.Name())
	out, err := json.MarshalIndent(data, "", "  ")
	check(err)
	_, err = p.WriteString(string(out) + "\n") //ioutil.WriteFile(p, out, os.FileMode(mode))
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
