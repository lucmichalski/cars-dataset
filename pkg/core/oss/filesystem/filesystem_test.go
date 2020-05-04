package filesystem

import (
	"testing"

	"github.com/lucmichalski/cars-dataset/pkg/core/oss/tests"
)

func TestAll(t *testing.T) {
	fileSystem := New("/tmp")
	tests.TestAll(fileSystem, t)
}
