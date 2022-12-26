package blogfs

import (
	"html/template"
	"io"
	"log"
	"os/exec"
)

var djotCmd *exec.Cmd

func DjotToHtml(djotText string) template.HTML {
	cmd := djotCommand()

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

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
