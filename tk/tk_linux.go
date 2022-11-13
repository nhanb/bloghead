//go:build linux

package tk

import _ "embed"

//go:embed vendored/tclkit-linux-amd64
var tclbin []byte
