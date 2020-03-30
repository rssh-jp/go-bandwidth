package bandwidth

import (
	"errors"
	"io"
	"log"
	"sync"
	"time"
)

var (
	// ErrCouldNotFoundFunction means function not defined
	ErrCouldNotFoundFunction = errors.New("Could not found function")
)

var (
	defaultVariable variable
	defaultConstant constant
)

func init() {
	defaultConstant = constant{
		limit:    1024 * 1024,
		duration: time.Second,
	}

	defaultVariable = variable{
		t: time.Now(),
	}
}

type constant struct {
	limit    int64
	duration time.Duration
}

type variable struct {
	bytes  int64
	t      time.Time
	tcount time.Duration
	sync.Mutex
}

// SetDefault is setting default constant
func SetDefault(limit int64, duration time.Duration) {
	defaultConstant.limit = limit
	defaultConstant.duration = duration
}

// ReadWriter represents extension structure of io.ReadWriter interface
type ReadWriter struct {
	fnRead   func(p []byte) (int, error)
	fnWrite  func(p []byte) (int, error)
	constant constant
	variable *variable
}

// NewReader returns ReadWriter structure included Read function
func NewReader(r io.Reader, limit int64, duration time.Duration) *ReadWriter {
	log.Println("#########", r)
	return &ReadWriter{
		fnRead: r.Read,
		constant: constant{
			limit:    limit,
			duration: duration,
		},
		variable: &variable{
			t: time.Now(),
		},
	}
}

// NewWriter returns ReadWriter structure included Write function
func NewWriter(w io.Writer, limit int64, duration time.Duration) *ReadWriter {
	return &ReadWriter{
		fnWrite: w.Write,
		constant: constant{
			limit:    limit,
			duration: duration,
		},
		variable: &variable{
			t: time.Now(),
		},
	}
}

// NewReadWriter returns ReadWriter structure included Read/Write function
func NewReadWriter(r io.Reader, w io.Writer, limit int64, duration time.Duration) *ReadWriter {
	return &ReadWriter{
		fnRead:  r.Read,
		fnWrite: w.Write,
		constant: constant{
			limit:    limit,
			duration: duration,
		},
		variable: &variable{
			t: time.Now(),
		},
	}
}

func NewReaderDefault(r io.Reader) *ReadWriter {
	return &ReadWriter{
		fnRead:   r.Read,
		constant: defaultConstant,
		variable: &defaultVariable,
	}
}

func NewWriterDefault(w io.Writer) *ReadWriter {
	return &ReadWriter{
		fnWrite:  w.Write,
		constant: defaultConstant,
		variable: &defaultVariable,
	}
}

func NewReadWriterDefault(r io.Reader, w io.Writer) *ReadWriter {
	return &ReadWriter{
		fnRead:   r.Read,
		fnWrite:  w.Write,
		constant: defaultConstant,
		variable: &defaultVariable,
	}
}

// Read is io.Reader.Read function
func (r *ReadWriter) Read(p []byte) (int, error) {
	if r.fnRead == nil {
		return 0, ErrCouldNotFoundFunction
	}

	r.variable.Lock()
	defer r.variable.Unlock()

	return exec(p, r.constant.limit, r.constant.duration, &r.variable.bytes, &r.variable.t, &r.variable.tcount, r.fnRead)
}

// Write is io.Writer.Write function
func (r *ReadWriter) Write(p []byte) (int, error) {
	if r.fnWrite == nil {
		return 0, ErrCouldNotFoundFunction
	}

	r.variable.Lock()
	defer r.variable.Unlock()

	return exec(p, r.constant.limit, r.constant.duration, &r.variable.bytes, &r.variable.t, &r.variable.tcount, r.fnWrite)
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

			n, err = fn(p[index : index+size])

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
