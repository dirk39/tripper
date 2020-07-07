package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestInputData(t *testing.T) {
	content := []byte("10")
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile
	if "10" != inputData("Insert an int number: ") {
		t.Errorf("inputData failed: %v", err)
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
}
