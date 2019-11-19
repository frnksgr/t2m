package t2m

import (
	"log"
	"time"
)

// tasklet is a function to be executed in a go routine
// tasklet execution can be stopped with struct{} send to the done channel
// the only output channel a tasklet might use is the logger
type tasklet func(l *log.Logger, done <-chan struct{})

// block until stopped
func sleep() tasklet {
	return func(l *log.Logger, done <-chan struct{}) {
		l.Println("Start none")
		<-done
		l.Println("End none")
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

// consume CPU until stopped
// p: cpu amount to be consumed e.g. 0.2 == 20%
func cpu(p float64) tasklet {
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

// fail after done
func fail() tasklet {
	return func(l *log.Logger, done <-chan struct{}) {
		l.Println("Start failing")
		<-done
		l.Panicln("End failing")
	}
}

// crash after done
func crash() tasklet {
	return func(l *log.Logger, done <-chan struct{}) {
		l.Println("Start crashing")
		<-done
		l.Fatalln("End crasing")
	}
}

func createTasklet(name string) tasklet {
	var t tasklet
	switch name {
	case "sleep":
		t = sleep()
	case "fail":
		t = fail()
	case "crash":
		t = crash()
	case "cpu":
		t = cpu(0.25) // 25% CPU
	case "ram":
		t = ram(1024 * 1024 * 100) // 100 MB RAM
	}
	return t
}

func (n *node) execTask() {
	if n.TaskName == "" {
		return
	}
	t := createTasklet(n.TaskName)

	done := make(chan struct{})
	go t(n.logger, done)
	timer := time.NewTimer(time.Duration(n.TaskDuration) * time.Millisecond)
	<-timer.C
	close(done)
}
