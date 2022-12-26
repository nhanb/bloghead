package tk

import (
	_ "embed"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var tclPath string

func init() {
	if runtime.GOOS == "windows" {
		executablePath, err := os.Executable()
		if err != nil {
			panic(err)
		}

		dir := filepath.Dir(executablePath)
		tclPath = filepath.Join(dir, "tclkit.exe")
		return
	}

	tclPath = "tclsh"
}

type Action string

const (
	ActionCreateFile Action = "create"
	ActionOpenFile          = "open"
	ActionCancel            = "cancel"
)

func stringToAction(s string) Action {
	a := Action(s)
	switch a {
	case ActionCreateFile:
		return a
	case ActionOpenFile:
		return a
	case ActionCancel:
		return a
	default:
		log.Fatalf("invalid action string: %s", s)
		return ""
	}
}

//go:embed choose-action.tcl
var chooseActionScript string

// Shows a window asking user to create new or open existing .bloghead file.
// Returns resulted action and, in case of "create" or "open", the full path of
// the selected file.
func ChooseAction() (action Action, filePath string) {
	result := strings.TrimSpace(execTcl(chooseActionScript))
	// This tcl script prints "<action><space><filepath>".
	// If user closes the window, it prints nothing.

	if result == "" {
		return ActionCancel, ""
	}

	parts := strings.SplitN(result, " ", 2)
	if len(parts) != 2 {
		log.Fatalf("Bogus ChooseAction script output:\n---\n%s\n---\n", result)
	}

	return stringToAction(parts[0]), parts[1]
}

// Executes tcl script, returns stdout
func execTcl(script string) string {
	cmd := exec.Command(tclPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	if runtime.GOOS == "windows" {
		// By default this window is not focused and not even brought to
		// foreground on Windows. I suspect it's because tcl is exec'ed from
		// bloghead.exe. Minimizing then re-opening it seems to do the trick.
		// This workaround, however, makes the window unfocused on KDE, so
		// let's only use it on Windows.
		script += "\nwm iconify .\nwm deiconify .\n"
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, script)
	}()

	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(out))
}
