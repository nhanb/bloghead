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

const ActionCreateFile Action = "create"
const ActionOpenFile Action = "open"
const ActionCancel Action = "cancel"

//go:embed scripts/choose-action.tcl
var chooseActionScript string

//go:embed scripts/get-open-file.tcl
var openFileScript string

//go:embed scripts/get-save-file.tcl
var createFileScript string

// Shows a window asking user to create new or open existing .bloghead file.
func ChooseAction() Action {
	action := Action(strings.TrimSpace(execTcl(chooseActionScript)))
	switch action {
	case ActionCreateFile:
	case ActionOpenFile:
	case ActionCancel:
	default:
		log.Fatalf("Invalid action: %s", action)
	}
	return action
}

func OpenFile() string {
	return execTcl(openFileScript)
}

func CreateFile() string {
	return execTcl(createFileScript)
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
