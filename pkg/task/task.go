package task

import (
	"log"
	"time"
)

// Tasklet bla bla
type Tasklet func(l *log.Logger, done <-chan struct{})

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
