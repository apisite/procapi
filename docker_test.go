// +build docker

package pgcall

// WIP: base for test with internal docker starts
// code based on https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859

import (
	"os/exec"
	"testing"
)

var testHasDocker bool

func init() {
	if _, err := exec.LookPath("docker"); err == nil {
		testHasDocker = true
	}
}
func TestDocker(t *testing.T) {
	if !testHasDocker {
		t.Log("docker not found, skipping")
		t.Skip()
	}
	// ...

	/*
	   cmd := exec.Command(os.Args[0], "-test.run=TestFailingGit")
	    cmd.Env = append(os.Environ(), "BE_CRASHING_GIT=1")
	    err := cmd.Run()
	    if e, ok := err.(*exec.ExitError); ok && !e.Success() {
	      return
	    }
	    t.Fatalf("Process ran with err %v, want os.Exit(1)", err)
	*/
}
