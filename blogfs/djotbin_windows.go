//go:build windows

package blogfs

import _ "embed"

//go:embed djotbin.exe
var djotbin []byte
