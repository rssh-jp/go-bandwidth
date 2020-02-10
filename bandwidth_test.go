package bandwidth

import (
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

func TestWriterSuccess(t *testing.T) {
	fd, err := os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}

	defer fd.Close()

	w := NewWriter(fd, limit, duration)

	for i := 0; i < 100; i++ {
		_, err := w.Write([]byte{2})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestReaderSuccess(t *testing.T) {
	fd, err := os.Open(filepath)
	if err != nil {
		t.Fatal(err)
	}

	defer fd.Close()

	r := NewReader(fd, limit, duration)

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

func TestReadWriterSuccess(t *testing.T) {
	fd, err := os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}

	defer fd.Close()

	rw := NewReadWriter(fd, fd, limit, duration)

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

func TestPostProcess(t *testing.T) {
	err := os.Remove(filepath)
	if err != nil {
		t.Fatal(err)
	}
}
