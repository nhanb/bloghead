// Must run CreateTclBin() before doing anything with this package.
package tk

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

var tmpTclPath string

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

//go:embed scripts/choose-action.tcl
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
	cmd := exec.Command(tmpTclPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
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

// Writes the embeded tcl executable into a temp file, overwriting any
// existing file to make sure we always have the most up-to-date version.
//
// Returns a cleanup function that caller must then call on shutdown.
func CreateTclBin() (cleanup func()) {
	tmpTclPath = path.Join(os.TempDir(), "bloghead-tcl")
	if runtime.GOOS == "windows" {
		// Windows wouldn't let me exec a file without an exe extension
		tmpTclPath += ".exe"
	}

	os.Remove(tmpTclPath)
	tmpFile, err := os.Create(tmpTclPath)
	if err != nil {
		panic(err)
	}
	defer tmpFile.Close()
	_, err = tmpFile.Write(tclbin)
	if err != nil {
		panic(err)
	}
	err = tmpFile.Chmod(fs.FileMode(0700))
	if err != nil {
		panic(err)
	}
	tmpTclPath = tmpFile.Name()
	fmt.Println("Created", tmpTclPath)

	return func() {
		os.Remove(tmpTclPath)
		fmt.Println("Removed", tmpTclPath)
	}
}
