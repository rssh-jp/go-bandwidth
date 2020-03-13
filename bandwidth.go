package bandwidth

import (
    "log"
	"errors"
	"io"
	"time"
)

var (
	// ErrCouldNotFoundFunction means function not defined
	ErrCouldNotFoundFunction = errors.New("Could not found function")
)

// ReadWriter represents extension structure of io.ReadWriter interface
type ReadWriter struct {
	limit    int64
	bytes    int64
	t        time.Time
	duration time.Duration
	fnRead   func(p []byte) (int, error)
	fnWrite  func(p []byte) (int, error)
    tcount   time.Duration
}

// NewReader returns ReadWriter structure included Read function
func NewReader(r io.Reader, limit int64, duration time.Duration) *ReadWriter {
    log.Println("#########", r)
	return &ReadWriter{
		limit:    limit,
		duration: duration,
		t:        time.Now(),
		fnRead:   r.Read,
	}
}

// NewWriter returns ReadWriter structure included Write function
func NewWriter(w io.Writer, limit int64, duration time.Duration) *ReadWriter {
	return &ReadWriter{
		limit:    limit,
		duration: duration,
		t:        time.Now(),
		fnWrite:  w.Write,
	}
}

// NewReadWriter returns ReadWriter structure included Read/Write function
func NewReadWriter(r io.Reader, w io.Writer, limit int64, duration time.Duration) *ReadWriter {
	return &ReadWriter{
		limit:    limit,
		duration: duration,
		t:        time.Now(),
		fnRead:   r.Read,
		fnWrite:  w.Write,
	}
}

// Read is io.Reader.Read function
func (r *ReadWriter) Read(p []byte) (int, error) {
	if r.fnRead == nil {
		return 0, ErrCouldNotFoundFunction
	}
	return exec(p, r.limit, r.duration, &r.bytes, &r.t, &r.tcount, r.fnRead)
}

// Write is io.Writer.Write function
func (r *ReadWriter) Write(p []byte) (int, error) {
	if r.fnWrite == nil {
		return 0, ErrCouldNotFoundFunction
	}
	return exec(p, r.limit, r.duration, &r.bytes, &r.t, &r.tcount, r.fnWrite)
}

const (
	taskCheck = iota + 1
	taskExec
	taskSleep
	taskEnd
)

func exec(p []byte, limit int64, duration time.Duration, bytes *int64, t *time.Time, tcount *time.Duration, fn func([]byte) (int, error)) (n int, err error) {
	task := taskCheck
	var index int64
	var retn int
	for isLoop := true; isLoop; {
		switch task {
		case taskCheck:
            s := time.Now()
			if index >= int64(len(p)) {
				task = taskEnd
				break
			}

			if *bytes >= limit {
				task = taskSleep
			} else {
				task = taskExec
			}
            *tcount += time.Now().Sub(s)
		case taskExec:
            s := time.Now()
			size := limit - *bytes
            l := int64(len(p[index:]))
			if size > l {
				size = l
			}

			n, err = fn(p[index:index+size])

			index += int64(n)
			retn += n
			*bytes += int64(n)

			if err != nil {
				task = taskEnd
				break
			}

			task = taskCheck

            *tcount += time.Now().Sub(s)
		case taskSleep:
			diff := duration - time.Now().Sub(*t)
            log.Println("SLEEP", diff, *tcount)
            *tcount = 0
			if diff > 0 {
				time.Sleep(diff)
			}

			*t = time.Now()

			*bytes -= limit

			task = taskCheck
		case taskEnd:
			isLoop = false
		}
	}

	return retn, err
}
