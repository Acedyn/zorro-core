//go:build js && wasm

package fs

import (
	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/indexeddb"
)

func init() {
	AvailableFileSystems()[config_proto.FileSystemType_IndexedDb] = func(config any) (hackpadfs.FS, error) { return indexeddb.NewFS(config), nil }
}
