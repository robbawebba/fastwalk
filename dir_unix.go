// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package fastwalk

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// import (
// 	"io"
// 	"os"
// 	"runtime"
// 	"syscall"
// )
//
const (
	blockSize = 4096 // TODO: calculate page size instead
)

//
func (info *INode) readdir(path string) ([]*INode, error) {
	f, err := os.Open(path) // consider syscall.Open for just getting fd
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var nodes []*INode
	var dirent *syscall.Dirent

	fd := int(f.Fd())

	buf := make([]byte, blockSize)
	for {
		buflen, err := syscall.ReadDirent(fd, buf)
		runtime.KeepAlive(f) // see KeepAlive godoc for an explanation
		if err != nil {
			return nil, err
		}

		if buflen <= 0 { // nothing to read
			break
		}

		for len(buf[:buflen]) > 0 { // this might not be a safe way of accessing the buffer
			// This stuff might be safer? (don't pay any attention to all the "unsafe" uses)
			/*
				      reclenOffset := unsafe.Offsetof(dirent.Reclen)
							reclenSize := unsafe.Sizeof(dirent.Reclen)

							reclen, _ := binary.Varint(buf[reclenOffset:reclenSize])
							if n != reclenSize {
								// error? reclen did not consume all of reclen.size (we we're expecting a full int)
							}
			*/
			dirent = (*syscall.Dirent)(unsafe.Pointer(&buf[0])) // point entry to first syscall.Dirent in buffer
			buf = buf[dirent.Reclen:]                           // reset buffer
			var node *INode
			switch dirent.Type {
			case syscall.DT_DIR:
				node.Mode = os.ModeDir
			case syscall.DT_LNK:
				node.Mode = os.ModeSymlink
			case syscall.DT_CHR:
				node.Mode = os.ModeDevice | os.ModeCharDevice
			case syscall.DT_BLK:
				node.Mode = os.ModeDevice
			case syscall.DT_FIFO:
				node.Mode = os.ModeNamedPipe
			case syscall.DT_SOCK:
				node.Mode = os.ModeSocket
			case syscall.DT_REG:
			default:
				// handle default, probably just do os.Stat
			}
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// 	dirname := info.Name
// 	if dirname == "" {
// 		dirname = "."
// 	}
// 	names, err := info.Readdirnames(n)
// 	nodes = make([]INode, 0, len(names))
// 	for _, filename := range names {
// 		fip, lerr := os.Lstat(dirname + "/" + filename)
// 		if os.IsNotExist(lerr) {
// 			// File disappeared between readdir + stat.
// 			// Just treat it as if it didn't exist.
// 			continue
// 		}
// 		if lerr != nil {
// 			return nodes, lerr
// 		}
//     fi
// 		nodes = append(nodes, fip)
// 	}
// 	if len(fi) == 0 && err == nil && n > 0 {
// 		// Per File.Readdir, the slice must be non-empty or err
// 		// must be non-nil if n > 0.
// 		err = io.EOF
// 	}
// 	return fi, err
// }
//
// func (info *INode) readdirnames(n int) ([]string, error) {
// 	// If this file has no dirinfo, create one.
// 	if f.dirinfo == nil {
// 		f.dirinfo = new(dirInfo)
// 		// The buffer must be at least a block long.
// 		f.dirinfo.buf = make([]byte, blockSize)
// 	}
// 	d := f.dirinfo
//
// 	size := n
// 	if size <= 0 {
// 		size = 100
// 		n = -1
// 	}
//
// 	names = make([]string, 0, size) // Empty with room to grow.
// 	for n != 0 {
// 		// Refill the buffer if necessary
// 		if d.bufp >= d.nbuf {
// 			d.bufp = 0
// 			var errno error
// 			d.nbuf, errno = f.pfd.ReadDirent(d.buf)
// 			runtime.KeepAlive(f)
// 			if errno != nil {
// 				return names, wrapSyscallError("readdirent", errno)
// 			}
// 			if d.nbuf <= 0 {
// 				break // EOF
// 			}
// 		}
//
// 		// Drain the buffer
// 		var nb, nc int
// 		nb, nc, names = syscall.ParseDirent(d.buf[d.bufp:d.nbuf], n, names)
// 		d.bufp += nb
// 		n -= nc
// 	}
// 	if n >= 0 && len(names) == 0 {
// 		return names, io.EOF
// 	}
// 	return names, nil
// }
