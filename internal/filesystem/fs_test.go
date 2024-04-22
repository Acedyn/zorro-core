package filesystem

import (
	"testing"

	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
)

// Testing the file system getter with the in memory file system config twice with the same config
func TestGetMemoryFileSystem(t *testing.T) {
	memfs, err := GetFileSystem(&filesystem_proto.FileSystemSettings{
		FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
	})

	if err != nil {
		t.Errorf("An error occured while getting the in memory file system:\n\t%s", err)
		return
	}

	secondMemfs, err := GetFileSystem(&filesystem_proto.FileSystemSettings{
		FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
	})

	if err != nil {
		t.Errorf("An error occured while getting the second in memory file system:\n\t%s", err)
		return
	}

	// Since we passed the same repository config twice we expect to reveive the same FS instance
	if memfs != secondMemfs {
		t.Errorf("Two different file system were given with same repository config")
		return
	}
}

// Testing the file system getter with the in memory file system config twice with a different config
func TestGetOsFileSystem(t *testing.T) {
	memfs, err := GetFileSystem(&filesystem_proto.FileSystemSettings{
		FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
	})

	if err != nil {
		t.Errorf("An error occured while getting the in os system:\n\t%s", err)
		return
	}

	secondMemfs, err := GetFileSystem(&filesystem_proto.FileSystemSettings{
		FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{
			Memory: &filesystem_proto.MemoryFsSettings{
				Name: "dummy",
			},
		},
	})

	if err != nil {
		t.Errorf("An error occured while getting the second os file system:\n\t%s", err)
		return
	}

	// Since we passed a different repository config we expect to reveive two different FS instance
	if memfs == secondMemfs {
		t.Errorf("The same file system was given twice with a different repository config")
		return
	}
}
