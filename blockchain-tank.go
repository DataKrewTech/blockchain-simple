/* README

Written by Sumanta Bose, 29 Apr 2018

MUX server methods available are:
    http://localhost:port/

FLAGS are:
  -port int
        mux server listen port (default 8080)
*/

package main

import (
	// "fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"

	// "os"
	"flag"
	"fmt"
	"log"
	"time"

	// "math"
	"strconv"
	// "runtime"
	"net/http"
	// "math/rand"
	// "io/ioutil"
	// "encoding/hex"
	"encoding/gob"
	"encoding/json"

	// "encoding/json"
	// "crypto/sha256"

	"github.com/gorilla/mux"
	// "github.com/davecgh/go-spew/spew"
)

///// GLOBAL FLAGS & VARIABLES

var StartTime time.Time

var listenPort, totalLocs *int // listen port & total locations in the supply chain
var dataDir *string            // pathname of data directory to save IoT Data

type IoTDataPoint struct {
	Temperature string `json:"Temperature"`
	Humidity    string `json:"Humidity"`
	Sound       string `json:"Sound"`
	Gas         string `json:"Gas"`
	PIR         string `json:"PIR"`
	SerialNo    int
}

var IoTDataArray []IoTDataPoint // To be saved as gob file

/////

type Block struct { // An element of Blockchain
	Index             int
	Timestamp         string
	IoTDataPointEntry []IoTDataPoint
	PrevHash          string
	ThisHash          string
}

var Blockchain []Block

///// LIST OF FUNCTIONS

func init() {

	gob.Register(IoTDataPoint{})
	gob.Register(Block{})
	gob.Register(map[string]interface{}{})

	log.SetFlags(log.Lshortfile)

	log.Printf("Welcome to Sumanta's IoT Dashboard Server!")
	listenPort = flag.Int("port", 8085, "mux server listen port")
	dataDir = flag.String("dataDir", "data", "pathname of data directory to save IoT Data")
	flag.Parse()

	LoadIoTData() // load from existing files, if any

	StartTime = time.Now()
	StartTime = StartTime.AddDate(0, -6, 10) // random negative offset
}

func main() {
	log.Fatal(launchMUXServer())
}

func launchMUXServer() error { // launch MUX server
	mux := makeMUXRouter()
	log.Println("HTTP Server Listening on port:", *listenPort) // listenPort is a global flag
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(*listenPort),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func makeMUXRouter() http.Handler { // create handlers
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleHome).Methods("GET")
	muxRouter.HandleFunc("/post/{Temperature}/{Humidity}/{Sound}/{Gas}/{PIR}", handlePost).Methods("GET")

	return muxRouter
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHome() API called")
	io.WriteString(w, "You have entered the restricted zone. Trespassing is strictly prohibited. Defaulters will be reported.")
}

func handlePost(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	newPoint := IoTDataPoint{}
	newPoint.SerialNo = len(IoTDataArray) + 1
	newPoint.Temperature = params["Temperature"]
	newPoint.Humidity = params["Humidity"]
	newPoint.Sound = params["Sound"]
	newPoint.Gas = params["Gas"]
	newPoint.PIR = params["PIR"]

	fmt.Println("Adding to SensorData:", newPoint)
	IoTDataArray = append(IoTDataArray, newPoint)
	gobCheck(writeIoTGob(IoTDataArray, len(IoTDataArray)))
	respondWithJSON(w, r, http.StatusCreated, newPoint)
}

func gobCheck(e error) { // Inspired from http://www.robotamer.com/code/go/gotamer/gob.html
	if e != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Println(line, "\t", file, "\n", e)
		os.Exit(1)
	}
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func writeIoTGob(object interface{}, fileNoCount int) error {
	filePath := *dataDir + "/IoT-data-" + strconv.Itoa(fileNoCount) + ".gob"
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func LoadIoTData() { // load from existing files, if any
	if _, err := os.Stat(*dataDir); os.IsNotExist(err) { // if *dataDir does not exist
		log.Println("`", *dataDir, "` does not exist. Creating directory.")
		os.Mkdir(*dataDir, 0755) // https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permission-value
	}

	files, err := ioutil.ReadDir(*dataDir) // dataDir from flag
	if err != nil {
		log.Fatal(err)
	}

	mostRecectFileNo := 0

	for _, file := range files {
		fileNo, _ := strconv.Atoi(file.Name()[len("IoT-data")+1 : len(file.Name())-4])
		if fileNo > mostRecectFileNo {
			mostRecectFileNo = fileNo
		}
	}

	if mostRecectFileNo == 0 {
		log.Println("No existing IoTData")
	} else {
		mostRecentFile := *dataDir + "/IoT-data-" + strconv.Itoa(mostRecectFileNo) + ".gob"
		log.Println("Loading existing IoTData from", mostRecentFile)
		gobCheck(readGob(&IoTDataArray, mostRecentFile))
	}
}

func readGob(object interface{}, filePath string) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}
