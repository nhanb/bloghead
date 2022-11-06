package blogfs

import (
	_ "embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
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

// Writes the embeded djotbin executable into a temp file, overwriting any
// existing file to make sure we always have the most up-to-date version.
//
// I would have preferred to os.CreateTemp() on startup and delete it on exit,
// but I haven't figured out how to run cleanups after exiting an http server,
// so this will have to do for now.
func CreateDjotbin() {
	tmpDjotbinPath = path.Join(os.TempDir(), "djotbin")
	os.Remove(tmpDjotbinPath)
	tmpFile, err := os.Create(tmpDjotbinPath)
	check(err)
	defer tmpFile.Close()
	_, err = tmpFile.Write(djotbin)
	check(err)
	err = tmpFile.Chmod(fs.FileMode(0700))
	check(err)
	tmpDjotbinPath = tmpFile.Name()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
