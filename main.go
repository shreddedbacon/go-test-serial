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
  I2CSlot        int  `json:"i2cslot"`
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

// Define a slot message type
type SlotMessage struct {
	Message string `json:"message"`
	Slots   []struct {
		I2CAddress int `json:"i2ca"`
		I2CSlot int `json:"i2cs"`
		PowerStatus   int `json:"ps"`
		AlwaysOn   int `json:"ao"`
		CaddyType   int `json:"ct"`
	} `json:"slots"`
}

var bushwodServer = ""
var bushwoodToken = ""

func main() {
  /* grab the details for the bushwod server so we can talk to it */
  bushwodServer = os.Getenv("BUSHWOOD_SERVER")
  bushwoodToken = os.Getenv("BUSHWOOD_TOKEN")
  serialDevice := os.Getenv("SERIAL_DEVICE")
  serialDeviceBaud := os.Getenv("SERIAL_DEVICE_BAUD")
  if bushwodServer == "" {
    log.Fatalln("BUSHWOOD_SERVER env var not set")
  }
  if bushwoodToken == "" {
    log.Fatalln("BUSHWOOD_TOKEN env var not set")
  }
  if serialDevice == "" {
    log.Fatalln("SERIAL_DEVICE env var not set")
  }
  if serialDeviceBaud == "" {
    log.Fatalln("SERIAL_DEVICE_BAUD env var not set")
  }

  /* create the serial manager to start the serial port communication */
  baudRate, _ := strconv.Atoi(serialDeviceBaud)
  serman, err := NewSerialManager(serialDevice, baudRate)
  if err != nil {
    log.Println(err)
  }

  /* start reading from port event handler */
	go readSer(serman.SerialPort, err)

  /* start the api */
  r := mux.NewRouter()
  r.HandleFunc("/api/v1/power/{i2cAddress}/{i2cSlot}/{powerStatus}", serman.sentToSer).Methods("GET")
  log.Println("Ready to check the backplanes!")
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", 8080), r))
  serman.SerialPort.Close()
}

/* function to send data to the serial port */
func (sm *serManager) sentToSer(w http.ResponseWriter, r *http.Request) {
  urlvars := mux.Vars(r)
  i2cAddress, _ := strconv.Atoi(urlvars["i2cAddress"])
  i2cSlot, _ := strconv.Atoi(urlvars["i2cSlot"])
  powerStatus, _ := strconv.Atoi(urlvars["powerStatus"])
  //i2cAddress := urlvars["i2cAddress"]
  //i2cSlot := urlvars["i2cSlot"]
  //powerStatus := urlvars["powerStatus"]
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
    log.Println(err)
  }
  log.Println("result write: "+string(result))
  //w.Write([]byte(string(result)+"\n"))
  json.NewEncoder(w).Encode(Acceptance{Success: fmt.Sprintf("sent command")})
}

/* create serial manager */
func NewSerialManager(serialDevice string, baudRate int) (*serManager, error) {
  c := &serial.Config{Name: serialDevice, Baud: baudRate, ReadTimeout: time.Millisecond * 50}
  s, err := serial.OpenPort(c)
  if err != nil {
    log.Println(err)
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
        if err.Error() != "EOF" {
          log.Println(err)
        }
      }
      if n == 0 {
        break
      }
      content = append(content, buf[:n]...)
    }
    /* check the content we get from the serial port */
    if len(content) != 0 {
      /* FIXME do better checking of the content to perform the right functions */
      contents := strings.Split(string(content), "\n")
      for b := range contents {
        if contents[b] != "" {
          log.Println(contents[b])
          slotMessage := SlotMessage{}
          if err := json.Unmarshal([]byte(strings.TrimSpace(string(contents[b]))), &slotMessage); err != nil {
            log.Println("Bad Message")
            log.Println(err)
          } else {
          	for _, v := range slotMessage.Slots {
          		slotAddress := SlotAddress{
          			I2CAddress: v.I2CAddress,
          			I2CSlot: v.I2CSlot,
          		}
              slotInfo := SlotInfo{
          			PowerStatus: v.PowerStatus,
          			AlwaysOn: v.AlwaysOn,
          			CaddyType: v.CaddyType,
          		}
              var netClient = &http.Client{
                Timeout: time.Second * 10,
              }
              token := bushwoodToken
              result, err2 := json.Marshal(slotInfo)
              if err2 != nil {
                log.Println(err2)
              }
              i2ca := strconv.Itoa(slotAddress.I2CAddress)
              i2cs := strconv.Itoa(slotAddress.I2CSlot)
              req, _ := http.NewRequest("POST", bushwodServer+"/api/v1/caddydata/i2c/" + i2ca + "/slot/" + i2cs, bytes.NewBuffer(result))
              req.Header.Add("apikey", token)
              req.Header.Set("Content-Type", "application/json")
              resp, resperr := netClient.Do(req)

              /* FIXME Check for connection! */

              if resperr != nil {
                log.Println(resperr)
              } else {
                if check200(resp.StatusCode) {
                  //fmt.Println("Successful update")
                  log.Println("Successful update")
                }
                resp.Body.Close()
              }
          	}
          }

          /* DELETE THIS BLOCK ----------->
          if strings.Contains(string(contents[b]), "{") {
            log.Println("result read: "+strings.TrimSpace(string(contents[b])))
            slotAddress := SlotAddress{}
            if err := json.Unmarshal([]byte(strings.TrimSpace(string(contents[b]))), &slotAddress); err != nil {
              //fmt.Println("err1")
              log.Println(err)
            }
            slotInfo := SlotInfo{}
            if err := json.Unmarshal([]byte(strings.TrimSpace(string(contents[b]))), &slotInfo); err != nil {
              //fmt.Println("err2")
              log.Println(err)
            }
            var netClient = &http.Client{
              Timeout: time.Second * 10,
            }
            token := bushwoodToken
            result, err2 := json.Marshal(slotInfo)
            if err2 != nil {
              log.Println(err2)
            }
            i2ca := strconv.Itoa(slotAddress.I2CAddress)
            i2cs := strconv.Itoa(slotAddress.I2CSlot)
            req, _ := http.NewRequest("POST", bushwodServer+"/api/v1/caddydata/i2c/" + i2ca + "/slot/" + i2cs, bytes.NewBuffer(result))
            req.Header.Add("apikey", token)
            req.Header.Set("Content-Type", "application/json")
            resp, resperr := netClient.Do(req)
            */
            /* FIXME Check for connection! */
            /*
            if resperr != nil {
              log.Println(resperr)
            } else {
              if check200(resp.StatusCode) {
                //fmt.Println("Successful update")
                log.Println("Successful update")
              }
              resp.Body.Close()
            }
          } else {
            log.Println("not json output: "+strings.TrimSpace(string(contents[b])))
          }
          <----------- DELETE THIS BLOCK */
        }
      }
    }
  }
}

func check200(httpcode int) bool {
	if httpcode == 200 {
		return true
	} else {
		return false
	}
}
