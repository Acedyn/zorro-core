package fs

import (
	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/os"
)

func init() {
	AvailableFileSystems()[config_proto.FileSystemType_Os] = func(any) (hackpadfs.FS, error) { return os.NewFS(), nil }
}
