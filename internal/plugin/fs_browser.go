//go:build js && wasm

package plugin

import (
	"io/fs"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	"github.com/hack-pad/hackpadfs/indexeddb"
)

func init() {
	AvailableFileSystems()[config_proto.FileSystemType_IndexedDb] = func(config any) (fs.FS, error) {
		switch config.(type) {
		case *config_proto.RepositoryConfig_IndexedDb:
			return indexeddb.NewFS()
		}

		return nil, fmt.Errorf("invalid config type passed")
	}
}
