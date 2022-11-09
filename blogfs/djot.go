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
	"runtime"
)

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
// Caller must also call CleanupDjotbin() on shutdown.
func CreateDjotbin() {
	tmpDjotbinPath = path.Join(os.TempDir(), "bloghead-djotbin")
	if runtime.GOOS == "windows" {
		// Windows wouldn't let me exec a file without an exe extension
		tmpDjotbinPath += ".exe"
	}

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

// Remember to call this on shutdown!
func CleanupDjotbin() {
	os.Remove(tmpDjotbinPath)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
