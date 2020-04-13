package bandwidth

import (
	"log"
	"os"
	"testing"
	"time"
)

func getFileSize(fd *os.File) (int64, error) {
	current, err := fd.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}

	_, err = fd.Seek(0, os.SEEK_SET)
	if err != nil {
		return 0, err
	}

	size, err := fd.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}

	_, err = fd.Seek(current, os.SEEK_SET)

	return size, nil
}

const (
	filepath = "./test"
	limit    = 10
	duration = time.Second
)

func postProcess() {
	err := os.Remove(filepath)
	if err != nil {
		log.Fatal(err)
	}
}

// TestMain is Main
func TestMain(m *testing.M) {
	defer postProcess()

	SetDefault(limit, duration)

	m.Run()
}

// TestWriterSuccess is Writer test
func TestWriterSuccess(t *testing.T) {
	t.Parallel()
	fd, err := os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}

	defer fd.Close()

	w := New(OptionUseDefault(), OptionWriter(fd))

	for i := 0; i < 100; i++ {
		_, err := w.Write([]byte{2})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestReaderSuccess is Reader test
func TestReaderSuccess(t *testing.T) {
	t.Parallel()
	fd, err := os.Open(filepath)
	if err != nil {
		t.Fatal(err)
	}

	defer fd.Close()

	r := New(OptionUseDefault(), OptionReader(fd))

	size, err := getFileSize(fd)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, size)

	n, err := r.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(n, buf)
}

// TestReadWriterSuccess is ReadWriter test
func TestReadWriterSuccess(t *testing.T) {
	t.Parallel()
	fd, err := os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}

	defer fd.Close()

	rw := New(OptionUseDefault(), OptionReader(fd), OptionWriter(fd))

	for i := 0; i < 50; i++ {
		_, err := rw.Write([]byte{5})
		if err != nil {
			t.Fatal(err)
		}
	}

	size, err := getFileSize(fd)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, size)

	_, err = fd.Seek(0, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}

	n, err := rw.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(n, buf)
}
