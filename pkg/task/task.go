package task

import (
	"log"
	"syscall"
	"time"
)

// Tasklet bla bla
type Tasklet func(l *log.Logger, done <-chan struct{})

// TaskletGenerator bla bla
//type TaskletGenerator func(a ...interface{}) Tasklet

// None bla bla
func None() Tasklet {
	return func(l *log.Logger, done <-chan struct{}) {
		l.Println("Start None")
		<-done
		l.Println("End None")
	}
}

// keep CPU busy for about us micro sec
func cpuloop(us int) {
	// dynamically calibrate on process start?
	// statically calibrated on Intel(R) Xeon(R) CPU E5-2667 0 @ 2.90GHz
	const lc = int(3.3 * 1000 * 1000)
	count := (us * lc) / 1000
	x, y := 0, 1
	for i := 1; i < count; i++ {
		x, y = y, x
	}
}

// CPU bla bla
func CPU(p float64) Tasklet {
	return func(l *log.Logger, done <-chan struct{}) {
		l.Printf("Start CPU load %2.2f%%", p*100)
		for end := false; !end; {
			select {
			case <-done:
				end = true
			default: // this should run about 10 ms
				t := time.Now()
				cpuloop(int(p * 10000.0))
				time.Sleep(
					time.Duration(
						(1.0 - p) * float64(time.Since(t))))
			}
		}
		l.Println("End CPU load")
	}

}

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
