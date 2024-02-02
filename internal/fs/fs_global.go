package fs

import (
	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
)

func init() {
	AvailableFileSystems()[config_proto.FileSystemType_Memory] = func(any) (hackpadfs.FS, error) { return mem.NewFS() }
}
