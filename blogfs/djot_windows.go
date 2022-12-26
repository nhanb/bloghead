//go:build windows

package blogfs

import (
	"os"
	"os/exec"
	"path/filepath"
)

var luaPath, djotPath string

// On Windows, we distribute all dependencies in the same folder as
// the main executable, so let's look it up once on startup:
func init() {
	executablePath, err := os.Executable()
	check(err)

	dir := filepath.Dir(executablePath)
	println("Dir:", dir)
	luaPath = filepath.Join(dir, "wlua54.exe")
	djotPath = filepath.Join(dir, "djot.lua")
}

func djotCommand() *exec.Cmd {
	return exec.Command(luaPath, djotPath)
}
