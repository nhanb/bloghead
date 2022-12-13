//go:build !windows

package blogfs

import _ "embed"

//go:embed djotbin
var djotbin []byte
