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
  "strconv"

  "github.com/tarm/serial"
  "github.com/gorilla/mux"
)

type serManager struct {
  SerialPort *serial.Port
}

type SlotPower struct {
  I2CAddress        int  `json:"i2caddress"`
  I2CSlot        int  `json:"i2slot"`
  PowerStatus        int  `json:"powercon"`
}

type SlotAddress struct {
  I2CAddress        int  `json:"i2ca"`
  I2CSlot        int  `json:"i2cs"`
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

func main() {
  /* grab the details for the greenskeeper server so we can talk to it */
  greensKeeper = os.Getenv("GK_SERVER")
  greensKeeperToken = os.Getenv("GK_TOKEN")
  if greensKeeper == "" {
    log.Fatalln("GK_SERVER env var not set")
  }
  if greensKeeperToken == "" {
    log.Fatalln("GK_TOKEN env var not set")
  }

  /* create the serial manager to start the serial port communication */
  serman, err := NewSerialManager()
  if err != nil {
    fmt.Println(err)
  }

  /* start reading from port event handler */
	go readSer(serman.SerialPort, err)

  /* start the api */
  r := mux.NewRouter()
  r.HandleFunc("/api/v1/power/{i2cAddress}/{i2cSlot}/{powerStatus}", serman.sentToSer).Methods("GET")
  log.Println("Ready to serve consoles!")
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", 8585), r))
  serman.SerialPort.Close()
}

/* function to send data to the serial port */
func (sm *serManager) sentToSer(w http.ResponseWriter, r *http.Request) {
  urlvars := mux.Vars(r)
  i2cAddress, _ := strconv.Atoi(urlvars["i2cAddress"])
  i2cSlot, _ := strconv.Atoi(urlvars["i2cSlot"])
  powerStatus, _ := strconv.Atoi(urlvars["powerStatus"])
  jsonData := SlotPower{
    I2CAddress: i2cAddress,
    I2CSlot: i2cSlot,
    PowerStatus: powerStatus,
  }
  result, err2 := json.Marshal(jsonData)
  if err2 != nil {
    log.Println(err2)
  }
  _, err := sm.SerialPort.Write([]byte(string(result)+"\n"))
  if err != nil {
    //fmt.Println(err)
  }
  fmt.Println(string(result))
  //w.Write([]byte(string(result)+"\n"))
  json.NewEncoder(w).Encode(Acceptance{Success: fmt.Sprintf("sent command")})
}

/* create serial manager */
func NewSerialManager() (*serManager, error) {
  c := &serial.Config{Name: "/dev/ttyS0", Baud: 9600, ReadTimeout: time.Millisecond * 50}
  s, err := serial.OpenPort(c)
  if err != nil {
    fmt.Println(err)
  }
  newInv := &serManager{
    SerialPort: s,
  }
  return newInv, nil
}

/* function to read data from serial port */
func readSer(s *serial.Port, err error) {
  for {
    buf := make([]byte, 40)
    var content []byte
    for {
      n, err := s.Read(buf)
      if err != nil {
        //need to fix this so it stops spewing "EOF" to screen
        //fmt.Println(err)
      }
      if n == 0 {
        break
      }
      content = append(content, buf[:n]...)
    }
    /* check the content we get from the serial port */
    if len(content) != 0 {
      /* FIXME do better checking of the content to perform the right functions */
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
      i2ca := strconv.Itoa(slotAddress.I2CAddress)
      i2cs := strconv.Itoa(slotAddress.I2CSlot)
      req, _ := http.NewRequest("POST", greensKeeper+"/api/v1/caddydata/i2c/" + i2ca + "/slot/" + i2cs, bytes.NewBuffer(result))
      req.Header.Add("apikey", token)
      req.Header.Set("Content-Type", "application/json")
      resp, _ := netClient.Do(req)
      defer resp.Body.Close()
    }
  }
}
