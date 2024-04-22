//go:build js && wasm

package filesystem

import (
	"context"
	"fmt"
	"io/fs"

	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	"github.com/hack-pad/hackpadfs/indexeddb"
)

func init() {
	AvailableFileSystems()[filesystem_proto.FileSystemType_IndexedDb] = func(config any) (fs.FS, error) {
		switch indexeddbConfig := config.(type) {
		case *filesystem_proto.RepositoryConfig_IndexedDb:
			return indexeddb.NewFS(context.Background(), indexeddbConfig.IndexedDb.Name, indexeddb.Options{})
		}

		return nil, fmt.Errorf("invalid config type passed")
	}
}
