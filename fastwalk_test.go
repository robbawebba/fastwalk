package fastwalk

import (
	"path/filepath"
	"testing"
)

var (
	validRoot   = "testDirs"
	invalidRoot = " "
)

func TestFastwalk(t *testing.T) {
	actualChildCount := 0
	expectedChildCount := 10 //TODO: calculate this during runtime
	err := Fastwalk(validRoot, func(path string, info *INode, err error) error {
		if err != nil {
			return err
		}
		if !info.isDir {
			actualChildCount++
			return nil
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
	err := Fastwalk(invalidRoot, func(path string, info *INode, err error) error {
		return err
	})
	if err == nil {
		t.Errorf("Invalid root turned out to be valid afterall... InvalidRoot=\"%s\"", invalidRoot)
	}
}

func TestFastwalkSkipDir(t *testing.T) {
	err := Fastwalk(validRoot, func(path string, info *INode, err error) error {
		return filepath.SkipDir
	})
	if err != nil {
		if err == filepath.SkipDir {
			return
		}
		t.Error(`An unexpected error occurred: `, err.Error())
	}
}
