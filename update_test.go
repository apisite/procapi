package pgcall

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const TestUpdateEnv = "TEST_UPDATE"
const TestUpdateDir = "testdata-new"

func checkTestUpdate(file string, data interface{}) {

	if os.Getenv(TestUpdateEnv) == "" {
		return
	}
	if _, err := os.Stat(TestUpdateDir); os.IsNotExist(err) {
		err = os.Mkdir(TestUpdateDir, os.ModePerm)
		check(err)
	}
	p, err := ioutil.TempFile(TestUpdateDir, file+".")
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
