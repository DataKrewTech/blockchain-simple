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
    "io"
    // "os"
    // "fmt"
    "log"
    "flag"
    "time"
    // "math"
    "strconv"
    // "runtime"
    "net/http"
    // "math/rand"
    // "io/ioutil"
    // "encoding/hex"
    "encoding/gob"
    // "encoding/json"
    // "crypto/sha256"

    "github.com/gorilla/mux"
    // "github.com/davecgh/go-spew/spew"
)

///// GLOBAL FLAGS & VARIABLES

var StartTime time.Time

var listenPort, totalLocs *int // listen port & total locations in the supply chain
var dataDir *string // pathname of data directory to save IoT Data

type IoTDataPoint struct {
    Temperature string `json:"Temperature"`
    Humidity string `json:"Humidity"`
    Sound string `json:"Sound"`
    Gas string `json:"Gas"`
    PIR string `json:"PIR"`
}

var IoTDataArray []IoTDataPoint // To be saved as gob file

/////

type Block struct { // An element of Blockchain
    Index int
    Timestamp string
    IoTDataPointEntry []IoTDataPoint
    PrevHash string
    ThisHash string
}

var Blockchain []Block

///// LIST OF FUNCTIONS

func init() {

    gob.Register(IoTDataPoint{}) ; gob.Register(Block{}) ;
    gob.Register(map[string]interface{}{})

    log.SetFlags(log.Lshortfile)

    log.Printf("Welcome to Sumanta's IoT Dashboard Server!")
    listenPort = flag.Int("port", 8085, "mux server listen port")
    dataDir = flag.String("dataDir", "data", "pathname of data directory to save IoT Data")
    flag.Parse()

    // LoadProductData() // load from existing files, if any

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
 
    newProduct := Product{}
    newProduct.SerialNo = len(ProductData) + 1
    newProduct.Location = 1
    newProduct.Name = params["Name"]
    codeNo, err1 := strconv.Atoi(params["CodeNo"]) ; newProduct.CodeNo = codeNo
    cost, err2 := strconv.Atoi(params["Cost"]) ; newProduct.Cost = cost

    if err1 == nil && err2 == nil {
        fmt.Println("Adding to ProductData:", newProduct)
        ProductData = append(ProductData, newProduct)
        gobCheck(writePdtGob(ProductData, len(ProductData)))
        respondWithJSON(w, r, http.StatusCreated, newProduct)
    }    
}


