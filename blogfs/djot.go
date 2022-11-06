package blogfs

import (
	_ "embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
)

//go:embed djotbin
var djotbin []byte

var tmpDjotbinPath string

func djotToHtml(djotText string) template.HTML {
	cmd := exec.Command(tmpDjotbinPath)

	stdin, err := cmd.StdinPipe()
	check(err)

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, djotText)
	}()

	out, err := cmd.Output()
	check(err)

	return template.HTML(out)
}

func CreateDjotbin() {
	tmpFile, err := os.CreateTemp("", "djotbin")
	check(err)
	defer tmpFile.Close()
	_, err = tmpFile.Write(djotbin)
	check(err)
	err = tmpFile.Chmod(fs.FileMode(0700))
	check(err)
	tmpDjotbinPath = tmpFile.Name()
	println("Created", tmpDjotbinPath)
}

func DeleteDjotbin() {
	err := os.Remove(tmpDjotbinPath)
	check(err)
	println("Deleted", tmpDjotbinPath)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
