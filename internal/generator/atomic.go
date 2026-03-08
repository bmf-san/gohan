package generator

import (
	"os"
	"path/filepath"
)

// writeFileAtomic writes data to path atomically: it creates a temporary file
// in the same directory, writes the content, then renames it to path.
//
// Using os.Rename (which maps to rename(2) on POSIX) is atomic, so the
// HTTP file-server never reads a partially-written file. Without this,
// os.WriteFile would truncate the target file to 0 bytes before writing,
// causing the dev server to serve an empty response and the browser to show
// a white page during a rebuild.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".gohan-tmp-")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil {
		_ = os.Remove(tmpName)
		return werr
	}
	if cerr != nil {
		_ = os.Remove(tmpName)
		return cerr
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
