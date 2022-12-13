package common

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"runtime"
)

// Writes an embedded blob into a temp executable file.
// Returns full path of that file.
//
// If file already exists with correct size, do nothing.
//
// filename will be prefixed with "bloghead-", to avoid name clashing.
func EnsureExecutable(data []byte, filename string) (filePath string) {
	filePath = path.Join(os.TempDir(), "bloghead-"+filename)
	if runtime.GOOS == "windows" {
		// Windows wouldn't let me exec a file without an exe extension
		filePath += ".exe"
	}

	info, err := os.Stat(filePath)
	if err == nil && !info.IsDir() {
		// Since the file name is already prefixed with "bloghead-", it's
		// practically impossible for any other file to accidentally exist
		// under the same name _and_ size. Saves us some CPU cycles (and more
		// importantly, startup time) for skipping checksum.
		if info.Size() == int64(len(data)) {
			fmt.Println("Found existing", filePath)
			return
		}
	}

	os.Remove(filePath)
	tmpFile, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer tmpFile.Close()
	_, err = tmpFile.Write(data)
	if err != nil {
		panic(err)
	}
	err = tmpFile.Chmod(fs.FileMode(0700))
	if err != nil {
		panic(err)
	}
	fmt.Println("Created", tmpFile.Name())
	return filePath
}
