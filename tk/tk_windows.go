//go:build windows

package tk

import _ "embed"

//go:embed vendored/tclkit-windows-amd64.exe
var tclbin []byte
