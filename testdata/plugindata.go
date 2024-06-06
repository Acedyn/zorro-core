package testdata

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/Acedyn/zorro-core/internal/filesystem"
	"github.com/Acedyn/zorro-core/internal/utils"

	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	"github.com/hack-pad/hackpadfs"
)

//go:embed plugins
var testdataPlugins embed.FS

//go:embed configs
var testdataConfigs embed.FS

//go:embed processors
var testdataProcessors embed.FS

// Initialize the in memory file system so it can be tested on both desktop and browser
func init() {
	memfs, err := filesystem.GetFileSystem(&filesystem_proto.FileSystemSettings{
		FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
	})

	if err != nil {
		utils.Logger().Error("Could not initialize test data in mem fs:\n\t%s", err)
		return
	}

	for _, testdata := range []embed.FS{testdataPlugins, testdataConfigs, testdataProcessors} {
		// Copy all the files and folders one by one
		err = fs.WalkDir(testdata, ".", func(path string, f os.DirEntry, _ error) error {
			if f.IsDir() {
				hackpadfs.Mkdir(memfs, path, 777)
			} else {
				file, err := hackpadfs.Create(memfs, path)
				if err != nil {
					return fmt.Errorf("Could not copy %s: %w", path, err)
				}

				fileContent, err := testdata.ReadFile(path)
				if err != nil {
					return fmt.Errorf("Could not copy %s: %w", path, err)
				}

				hackpadfs.WriteFile(file, fileContent)
			}

			return nil
		})

		if err != nil {
			utils.Logger().Error("An error occured during embed fs copy:\n\t%s", err)
			return
		}
	}
}
