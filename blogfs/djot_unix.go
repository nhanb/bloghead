//go:build !windows

package blogfs

import "os/exec"

// On unixlikes we assume whatever installation method ensured that
// djot.lua is executable and available in $PATH.
func djotCommand() *exec.Cmd {
	return exec.Command("djot.lua")
}
