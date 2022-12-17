package blogfs

import (
	_ "embed"
	"html/template"
	"io"
	"log"
	"os/exec"

	"go.imnhan.com/bloghead/common"
)

var tmpDjotbinPath string

func DjotToHtml(djotText string) template.HTML {
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

func EnsureDjotBin() {
	tmpDjotbinPath = common.EnsureExecutable(djotbin, "djotbin")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
