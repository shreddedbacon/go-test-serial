package main

import (
	"fmt"
	"time"
  "net/http"
  "log"


	"github.com/tarm/serial"
  "github.com/gorilla/mux"
)

func main() {
  c := &serial.Config{Name: "/dev/ttyS0", Baud: 9600, ReadTimeout: time.Millisecond * 500}
  s, err := serial.OpenPort(c)
  if err != nil {
    fmt.Println(err)
  }

	go readSer(s, err)

  r := mux.NewRouter()
  r.HandleFunc("/api/v1/consoles", sentToSer).Methods("GET")
  log.Println("Ready to serve consoles!")
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", 8585), r))
  s.Close()
}

func sentToSer(w http.ResponseWriter, req *http.Request) {
  w.Write([]byte("return\n"))
}

func readSer(s *serial.Port, err error) {
  for {
    _, err = s.Write([]byte("bobtestbob\n"))
    if err != nil {
      //fmt.Println(err)
    }
    time.Sleep(time.Second / 2)

    buf := make([]byte, 40)
    for {
      n, err := s.Read(buf)

      if err != nil {
        //fmt.Println(err)
      }
      if n == 0 {
        //fmt.Println("\nEOF")
        break
      }
      //fmt.Println(string(buf[:n]))
    }
  }
}
