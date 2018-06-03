package main

import (
	"fmt"
	"time"
  "net/http"
  "log"
  "strings"
  "encoding/json"
  "os"
  "bytes"

	"github.com/tarm/serial"
  "github.com/gorilla/mux"
)

type serManager struct {
  ser *serial.Port
}

type SlotPower struct {
  I2CAddress        int  `json:"i2caddress"`
  I2CSlot        int  `json:"i2slot"`
  PowerStatus        int  `json:"powerstatus"`
}

type SlotAddress struct {
  I2CAddress        int  `json:"i2ca"`
  I2CSlot        int  `json:"i2s"`
}

type SlotInfo struct {
  PowerStatus        int  `json:"ps"`
  AlwaysOn        int  `json:"ao"`
  CaddyType        int  `json:"ct"`
}

type ExceptionE struct {
  Error string `json:"error"`
}

type Acceptance struct {
  Success string `json:"success"`
}

var greensKeeper = ""
var greensKeeperToken = ""

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
  greensKeeper = os.Getenv("GK_SERVER")
  greensKeeperToken = os.Getenv("GK_TOKEN")
  if greensKeeper == "" {
    log.Fatalln("GK_SERVER env var not set")
  }
  if greensKeeperToken == "" {
    log.Fatalln("GK_TOKEN env var not set")
  }

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
      slotAddress := SlotAddress{}
      if err := json.Unmarshal([]byte(content), &slotAddress); err != nil {
        log.Println(err)
      }
      slotInfo := SlotInfo{}
      if err := json.Unmarshal([]byte(content), &slotInfo); err != nil {
        log.Println(err)
      }
      var netClient = &http.Client{
        Timeout: time.Second * 10,
      }
      token := greensKeeperToken
      result, err2 := json.Marshal(slotInfo)
      if err2 != nil {
        log.Println(err2)
      }
      req, _ := http.NewRequest("POST", greensKeeper+"/api/v1/caddydata/i2c/" + slotAddress.I2CAddress + "/slot/" + slotAddress.I2CSlot, bytes.NewBuffer(result))
      //req.Header.Add("Authorization", "Bearer "+token)
      req.Header.Add("apikey", token)
      req.Header.Set("Content-Type", "application/json")
      resp, _ := netClient.Do(req)
      //body, _ := ioutil.ReadAll(resp.Body)
      //log.Printf(string(body))
      defer resp.Body.Close()
    }
  }
}
