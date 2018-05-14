package fastwalk

import (
	"os"
	"path/filepath"
)

// WalkFunc is the type of the function called for each file or directory
// visited by Walk. The path argument is the absolute path to a file or
// directory. For example, if info argument represents a file named "a" that
// exists in the directory "dir", then path will be "dir/a". The info argument
// is the INode for the named path.
//
// If there was a problem walking to the file or directory named by path, the
// incoming error will describe the problem and the function can decide how
// to handle that error (and Walk will not descend into that directory). If
// an error is returned, processing stops. The sole exception is when the function
// returns the special value filepath.SkipDir. If the function returns filepath.SkipDir when invoked
// on a directory, Walk skips the directory's contents entirely.
// If the function returns filepath.SkipDir when invoked on a non-directory file,
// Fastwalk skips the remaining files in the containing directory.
type WalkFunc func(path string, info *INode, err error) error

// Walk traverses the file tree depth-first starting at root, calling walkFn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. Unlike filepath.Walk, Fastwalk does
// not walk the files in lexical order and does not gather all information about
// a file (i.e. no os.FileInfo).
// Walk does not follow symbolic links.
func Walk(root string, walkFn WalkFunc) error {
	fi, err := os.Lstat(root)
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		rootINode := &INode{
			fi.Mode(),
			fi.Name(),
		}
		err = walk(root, rootINode, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

func walk(path string, info *INode, walkFn WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	nodes, err := readdir(path)
	err1 := walkFn(path, info, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means walkFn wants walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err != nil || err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by walkFn. walkFn may ignore err and return nil.
		// If walkFn returns filepath.SkipDir, it will be handled by the caller.
		// So walk should return whatever walkFn returns.
		return err1
	}

	for _, node := range nodes {
		fullpath := filepath.Join(path, node.Name)
		err = walk(fullpath, node, walkFn)
		if err != nil {
			if !node.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}

	return nil
}
