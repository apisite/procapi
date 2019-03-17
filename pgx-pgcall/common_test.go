package pgxpgcall

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {

	myTest := &ServerSuite{}
	suite.Run(t, myTest)

	myTest.hook.Reset()

	for _, e := range myTest.hook.Entries {
		fmt.Printf("ENT[%s]: %s\n", e.Level, e.Message)
	}

}

