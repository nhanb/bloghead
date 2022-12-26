package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type remoteFile struct {
	url       string
	destPath  string
	sha256sum string
}

var remoteFiles = map[string]remoteFile{
	"linux": {
		url:       "https://tclkits.rkeene.org/fossil/raw/tclkit-8.6.3-rhel5-x86_64?name=36b5cb68899cfcb79417a29f9c6d8176ebae0d24",
		destPath:  "vendored/tclkit/tclkit-linux-amd64",
		sha256sum: "dba225a4a3e1c2bfbae68d98b95f564fe14619eda83d1903116465a047bb2ca0",
	},
	"windows": {
		url:       "https://tclkits.rkeene.org/fossil/raw/tclkit-8.6.3-win32-x86_64.exe?name=403c507437d0b10035c7839f22f5bb806ec1f491",
		destPath:  "vendored/tclkit/tclkit.exe",
		sha256sum: "5292399891398ce13af0e32fa98dab02e6f0134ea9738515649d7e649eff0942",
	},
}

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "linux" && os.Args[1] != "windows") {
		log.Fatalf("Usage: go run ./cmd/vendortclkit linux|windows")
	}

	f := remoteFiles[os.Args[1]]
	fmt.Printf("Downloading tclkit to %s...\n", f.destPath)
	download(f.url, f.destPath)

	fmt.Print("SHA1 checksum... ")
	actualSum := checksum(f.destPath)
	if actualSum != f.sha256sum {
		log.Fatalf(
			"mismatched!\nExpected: %s\nGot: %s\n", f.sha256sum, actualSum,
		)
	}
	fmt.Println("matched.")
}

func download(from string, to string) {
	out, err := os.Create(to)
	check(err)
	defer out.Close()

	resp, err := http.Get(from)
	check(err)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	check(err)
}

func checksum(filePath string) string {
	f, err := os.Open(filePath)
	check(err)
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	check(err)

	return hex.EncodeToString(h.Sum(nil))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
