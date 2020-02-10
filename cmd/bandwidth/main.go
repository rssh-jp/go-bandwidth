package main

import (
	"log"
	"os"
	"time"

	"github.com/rssh-jp/go-bandwidth"
)

func main() {
	fd, err := os.OpenFile("./test", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		log.Fatal(err)
	}

	defer fd.Close()

	w := bandwidth.NewWriter(fd, 100, time.Second)

	for i := 0; i < 100; i++ {
		_, err := w.Write([]byte{2})
		if err != nil {
			log.Fatal(err)
		}
	}
}
