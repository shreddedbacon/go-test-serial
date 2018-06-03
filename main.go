package main

import (
	"fmt"
	"time"
  "net/http"
  "log"
  "strings"
  "encoding/json"

	"github.com/tarm/serial"
  "github.com/gorilla/mux"
)

type serManager struct {
  ser *serial.Port
}

type SlotPower struct {
  I2CAddress        string  `json:"i2caddress"`
  I2CSlot        string  `json:"i2slot"`
  PowerStatus        string  `json:"powerstatus"`
}

type ExceptionE struct {
  Error string `json:"error"`
}

type Acceptance struct {
  Success string `json:"success"`
}

func (sm *serManager) sentToSer(w http.ResponseWriter, r *http.Request) {
  urlvars := mux.Vars(r)
  jsonData := SlotPower{
    I2CAddress: urlvars["i2cAddress"],
    I2CSlot: urlvars["i2cSlot"],
    PowerStatus: urlvars["powerStatus"],
  }
  result, err2 := json.Marshal(jsonData)
  if err2 != nil {
    log.Println(err2)
  }
  _, err := sm.ser.Write([]byte(string(result)+"\n"))
  if err != nil {
    //fmt.Println(err)
  }
  //w.Write([]byte(string(result)+"\n"))
  json.NewEncoder(w).Encode(Acceptance{Success: fmt.Sprintf("sent command")})
}

/* create ser man */
func NewSerMan() (*serManager, error) {
  c := &serial.Config{Name: "/dev/ttyS0", Baud: 9600, ReadTimeout: time.Millisecond * 50}
  s, err := serial.OpenPort(c)
  if err != nil {
    fmt.Println(err)
  }
  newInv := &serManager{
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
  r.HandleFunc("/api/v1/power/{i2cAddress}/{i2cSlot}/{powerStatus}", serman.sentToSer).Methods("GET")
  log.Println("Ready to serve consoles!")
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", 8585), r))
  serman.ser.Close()
}

func readSer(s *serial.Port, err error) {
  for {
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
      fmt.Println(strings.TrimSpace(string(content)))
    }
  }
}
