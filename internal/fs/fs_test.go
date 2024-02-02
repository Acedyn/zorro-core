package fs

import (
	"testing"
)

// Test file system registration
func TestAvailableFileSystems(t *testing.T) {
	availableFileSystems := AvailableFileSystems()
	expectedFileSystemsCount := 2
	availableFileSystemsCount := len(availableFileSystems)

	if availableFileSystemsCount < expectedFileSystemsCount {
		t.Errorf("Missing available file systems (registered %d, expexted %d)", availableFileSystemsCount, expectedFileSystemsCount)
	}

	if availableFileSystemsCount > expectedFileSystemsCount {
		t.Errorf("Too many available file systems (registered %d, expexted %d)", availableFileSystemsCount, expectedFileSystemsCount)
	}
}
