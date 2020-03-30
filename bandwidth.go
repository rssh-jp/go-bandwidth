package bandwidth

import (
	"errors"
	"io"
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

// Option is functional option pattern
type Option func(*ReadWriter)

// OptionReader returns Option instance containing reader setting
func OptionReader(r io.Reader) Option {
	return func(rw *ReadWriter) {
		rw.fnRead = r.Read
	}
}

// OptionWriter returns Option instance containing writer setting
func OptionWriter(w io.Writer) Option {
	return func(rw *ReadWriter) {
		rw.fnWrite = w.Write
	}
}

// OptionConstant returns Option instance containing constant setting
func OptionConstant(limit int64, duration time.Duration) Option {
	return func(rw *ReadWriter) {
		rw.constant.limit = limit
		rw.constant.duration = duration
	}
}

// OptionUseDefault returns Option instance containing use default setting
func OptionUseDefault() Option {
	return func(rw *ReadWriter) {
		rw.constant = defaultConstant
		rw.variable = &defaultVariable
	}
}

// New returns ReadWriter instance
func New(options ...Option) *ReadWriter {
	rw := &ReadWriter{
		variable: &variable{
			t: time.Now(),
		},
	}

	for _, opt := range options {
		opt(rw)
	}

	return rw
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
