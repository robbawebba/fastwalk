// +build linux darwin freebsd netbsd openbsd

package fastwalk

import (
	"bytes"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	blockSize = 4096 // TODO: calculate block size instead
)

//
func readdir(path string) ([]*INode, error) {
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
		if err != nil {
			return nil, err
		}
		runtime.KeepAlive(f) // see KeepAlive godoc for an explanation

		if buflen <= 0 { // nothing to read
			break
		}

		filledBuf := buf[:buflen]
		for len(filledBuf) > 0 {
			// this might not be a safe way of accessing the buffer
			// This stuff might be safer? (don't pay any attention to all the "unsafe" uses)
			// /*
			// 	      reclenOffset := unsafe.Offsetof(dirent.Reclen)
			// 				reclenSize := unsafe.Sizeof(dirent.Reclen)
			//
			// 				reclen, _ := binary.Varint(buf[reclenOffset:reclenSize])
			// 				if n != reclenSize {
			// 					// error? reclen did not consume all of reclen.size (we we're expecting a full int)
			// 				}
			// */
			dirent = (*syscall.Dirent)(unsafe.Pointer(&filledBuf[0])) // point entry to first syscall.Dirent in buffer
			filledBuf = filledBuf[dirent.Reclen:]                     // reset buffer
			node := &INode{}
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
				// a regular file. node.Mode&os.ModeType will be 0
			default:
				// handle default, probably just do os.Stat
			}

			nameBuf := (*[unsafe.Sizeof(dirent.Name)]byte)(unsafe.Pointer(&dirent.Name[0]))
			nameLen := bytes.IndexByte(nameBuf[:], 0)
			if nameLen < 0 {
				panic("failed to find terminating 0 byte in dirent")
			}
			// Special cases for `.`` & `..` entries:
			if nameLen == 1 && nameBuf[0] == '.' || nameLen == 2 && nameBuf[0] == '.' && nameBuf[1] == '.' {
				continue
			}

			node.Name = string(nameBuf[:nameLen])
			nodes = append(nodes, node)
		}
	}

	if err = f.Close(); err != nil {
		return nil, err
	}

	return nodes, nil
}
