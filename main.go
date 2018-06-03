package main

import (
	"fmt"
	"time"
  "net/http"
  "log"


	"github.com/tarm/serial"
  "github.com/gorilla/mux"
)

type appContext struct {
  ser *serial.Port
}

func (ah *appContext) sentToSer(w http.ResponseWriter, r *http.Request) {
  _, err := ah.ser.Write([]byte("bobtestbob\n"))
  if err != nil {
    //fmt.Println(err)
  }
  w.Write([]byte("return\n"))
}

/* create ser man */
func NewSerMan() (*appContext, error) {
  c := &serial.Config{Name: "/dev/ttyS0", Baud: 9600, ReadTimeout: time.Millisecond * 500}
  s, err := serial.OpenPort(c)
  if err != nil {
    fmt.Println(err)
  }
  newInv := &appContext{
    ser:                 s,
  }
  return newInv, nil
}

func main() {

  serman, err := NewSerMan()
  if err != nil {
    fmt.Println(err)
  }

	go readSer(serman.ser, err)

  r := mux.NewRouter()
  r.HandleFunc("/api/v1/consoles", serman.sentToSer).Methods("GET")
  log.Println("Ready to serve consoles!")
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", 8585), r))
  serman.ser.Close()
}

func readSer(s *serial.Port, err error) {
  for {
    /*_, err = s.Write([]byte("bobtestbob\n"))
    if err != nil {
      //fmt.Println(err)
    }
    time.Sleep(time.Second / 2)*/

    buf := make([]byte, 40)
    var content []byte
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
      content = append(content, buf[:n]...)
      //fmt.Println(n)
    }
    //fmt.Println(string(content))
    if len(content) != 0 {
      fmt.Println(string(content))
    }
  }
}
