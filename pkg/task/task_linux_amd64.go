// +build linux
// +build amd64

package task

import (
	"log"
	"syscall"
	"time"
)

func alloc(bytes int) ([]byte, error) {
	mem, err := syscall.Mmap(-1, 0, bytes,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANONYMOUS|syscall.MAP_PRIVATE)
	if err != nil {
		return nil, err
	}
	return mem, nil
}

func touch(mem []byte) {
	inc := syscall.Getpagesize() * 100 / 125
	for i := 0; i < len(mem); i += inc {
		mem[i] = 1
	}
}

func free(mem []byte) error {
	return syscall.Munmap(mem)
}

// RAM bla bla
func RAM(s uint64) Tasklet {
	return func(l *log.Logger, done <-chan struct{}) {
		l.Printf("Consume %d bytes of RAM\n", s)
		mem, err := alloc(int(s))
		if err != nil {
			l.Panic(err)
		}
		defer free(mem)
		for end := false; !end; {
			select {
			case <-done:
				end = true
			default:
				touch(mem)
				time.Sleep(time.Millisecond)
			}
		}
		l.Println("Free RAM")
	}
}
