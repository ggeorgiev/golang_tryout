package pack_test

import (
	"pack"
	"testing"
	"os"
	"log"
)

func TestExport(t *testing.T) {
	if pack.Export() == 0 {

	}
}
func TestExport2(t *testing.T) {
	_, err := os.Open("filename.ext")
	if err != nil {
		log.Fatal(err)
	}
}
