package fastwalk

import (
	"os"
	"path/filepath"
)

type INode struct {
	isDir bool
	name  string
}

type WalkFunc func(path string, info *INode, err error) error

func Fastwalk(root string, walkFunc WalkFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		err = walkFunc(root, nil, err)
	} else {
		rootINode := &INode{
			info.IsDir(),
			info.Name(),
		}
		err = walk(root, rootINode, walkFunc)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

func walk(path string, info *INode, walkFn WalkFunc) error {
	if !info.isDir {
		return walkFn(path, info, nil)
	}

	names, err := readDirNames(path)
	err1 := walkFn(path, info, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means walkFn wants walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err != nil || err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by walkFn. walkFn may ignore err and return nil.
		// If walkFn returns SkipDir, it will be handled by the caller.
		// So walk should return whatever walkFn returns.
		return err1
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, info, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(filename, info, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}

	return nil
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	// sort.Strings(names) // remove, no need to sort
	return names, nil
}
