package imagehelper

import (
	"encoding/base64"
	"os"
)

func saveimage() {
	var b64 string
	dec, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("filename")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.Write(dec); err != nil {
		panic(err)
	}
	if err := f.Sync(); err != nil {
		panic(err)
	}
}
