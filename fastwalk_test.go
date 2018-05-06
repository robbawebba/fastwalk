package fastwalk

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/karrick/godirwalk"
)

var (
	validRoot   = "testDirs"
	invalidRoot = " "
	benchDir    = flag.String("benchdir", runtime.GOROOT(), "The directory to walk for benchmarking testing")
	testErr     = errors.New(`this is a test!`)
)

func TestFastwalk(t *testing.T) {
	actualChildCount := 0
	expectedChildCount := 10 //TODO: calculate this during runtime
	err := Walk(validRoot, func(path string, info *INode, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			actualChildCount++
		}
		return nil
	})
	if err != nil {
		t.Error(`Error while walking validRoot: `, err.Error())
		return
	}
	if actualChildCount != expectedChildCount {
		t.Errorf(`actualChildCount does not match expectedChildCount (%d != %d)`, actualChildCount, expectedChildCount)
	}
}

func TestFastwalkInvalidRoot(t *testing.T) {
	err := Walk(invalidRoot, func(path string, info *INode, err error) error {
		return err
	})
	if err == nil {
		t.Errorf("Invalid root turned out to be valid afterall... InvalidRoot=\"%s\"", invalidRoot)
	}
}

func TestFastwalkSkipDir(t *testing.T) {
	err := Walk(validRoot, func(path string, info *INode, err error) error {
		return filepath.SkipDir
	})
	if err != nil {
		if err == filepath.SkipDir {
			return
		}
		t.Error(`An unexpected error occurred: `, err.Error())
	}
}
func TestFastwalkFileError(t *testing.T) {
	err := Walk(validRoot, func(path string, info *INode, err error) error {
		if info.IsDir() {
			return nil
		}
		return testErr
	})
	if err != testErr {
		t.Error(`An unexpected error occurred: `, err.Error())
	}
}

func BenchmarkFilepathWalk(b *testing.B) {
	var fileCount, dirCount int

	benchmarkWalkFunc := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			dirCount++
		} else {
			fileCount++
		}
		return nil
	}

	for n := 0; n < b.N; n++ {
		fileCount, dirCount = 0, 0
		err := filepath.Walk(*benchDir, benchmarkWalkFunc)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

func BenchmarkFastwalk(b *testing.B) {
	var fileCount, dirCount int

	benchmarkWalkFunc := func(path string, info *INode, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
		}
		return nil
	}

	for n := 0; n < b.N; n++ {
		fileCount, dirCount = 0, 0
		err := Walk(*benchDir, benchmarkWalkFunc)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

func BenchmarkFastwalkWithFileStat(b *testing.B) {
	var fileCount, dirCount int

	benchmarkWalkFunc := func(path string, info *INode, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirCount++
		} else {
			_, err := os.Stat(filepath.Join(path, info.Name))
			if err != nil {
				return err
			}
			fileCount++
		}
		return nil
	}

	for n := 0; n < b.N; n++ {
		fileCount, dirCount = 0, 0
		err := Walk(*benchDir, benchmarkWalkFunc)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

func BenchmarkGoDirWalk(b *testing.B) {
	var fileCount, dirCount int

	benchmarkWalkFunc := func(path string, de *godirwalk.Dirent) error {
		if de.IsDir() {
			dirCount++
		} else {
			fileCount++
		}
		return nil
	}

	for n := 0; n < b.N; n++ {
		fileCount, dirCount = 0, 0
		godirwalk.Walk(*benchDir, &godirwalk.Options{
			Callback: benchmarkWalkFunc,
		})
	}
}
