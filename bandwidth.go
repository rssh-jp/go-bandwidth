package bandwidth

import (
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
}

// NewReader returns ReadWriter structure included Read function
func NewReader(r io.Reader, limit int64, duration time.Duration) *ReadWriter {
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
	return r.exec(p, r.fnRead)
}

// Write is io.Writer.Write function
func (r *ReadWriter) Write(p []byte) (int, error) {
	if r.fnWrite == nil {
		return 0, ErrCouldNotFoundFunction
	}
	return r.exec(p, r.fnWrite)
}

const (
	taskCheck = iota + 1
	taskExec
	taskSleep
	taskEnd
)

func (r *ReadWriter) exec(p []byte, fn func([]byte) (int, error)) (n int, err error) {
	task := taskCheck
	var index int64
	var retn int
	for isLoop := true; isLoop; {
		switch task {
		case taskCheck:
			if index >= int64(len(p)) {
				task = taskEnd
				break
			}

			if r.bytes >= r.limit {
				task = taskSleep
			} else {
				task = taskExec
			}
		case taskExec:
			size := r.limit - r.bytes
			if size > int64(len(p[index:])) {
				size = int64(len(p[index:]))
			}

			b := p[index : index+size]

			n, err = fn(b)

			index += int64(n)
			retn += n
			r.bytes += int64(n)

			if err != nil {
				task = taskEnd
				break
			}

			task = taskCheck

		case taskSleep:
			diff := r.duration - time.Now().Sub(r.t)
			if diff > 0 {
				time.Sleep(diff)
			}

			r.t = time.Now()

			r.bytes -= r.limit

			task = taskCheck
		case taskEnd:
			isLoop = false
		}
	}

	return retn, err
}
