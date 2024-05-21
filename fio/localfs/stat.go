package localfs

import (
	"crypto/sha1"
	"encoding/hex"
	"io/fs"
	"log"
	"os"

	"github.com/NightmareZero/nzgoutil/fio"
)

var _ fio.IFileStat = &FileStatHandler{}

type FileStatHandler struct {
	name string
	fs.FileInfo
}

// Sha1 implements fio.IFileStat.
func (f *FileStatHandler) Sha1() string {
	data, err := os.ReadFile(f.name)
	if err != nil {
		log.Fatal(err)
	}

	hasher := sha1.New()
	hasher.Write(data)
	sha := hasher.Sum(nil)

	return hex.EncodeToString(sha)
}
