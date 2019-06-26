/* README

Written by Sumanta Bose, 25 June 2019

MUX server methods available are:
    http://localhost:port/
    http://localhost:port/post
    http://localhost:port/blocklist
    http://localhost:port/blockinfo/{Index}
    http://localhost:port/blockchain

FLAGS are:
  -port int
        mux server listen port (default 8080)
*/

package main

import (
	// "fmt"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"

	// "os"
	"flag"
	"fmt"
	"log"
	"time"

	"bytes"
	"strings"

	// "math"
	"strconv"
	// "runtime"
	"net/http"
	// "math/rand"
	// "io/ioutil"
	// "encoding/hex"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	gonet "net"

	// "crypto/sha256"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	// "github.com/davecgh/go-spew/spew"
)

///// GLOBAL FLAGS & VARIABLES

var StartTime time.Time

var listenPort, totalLocs *int // listen port & total locations in the supply chain
var dataDir *string            // pathname of data directory to save IoT Data

type Accelerometer struct {
	Ax string `json:"ax"`
	Ay string `json:"ay"`
	Az string `json:"az"`
}

type Gyroscope struct {
	Gx string `json:"gx"`
	Gy string `json:"gy"`
	Gz string `json:"gz"`
}

type Temperature struct {
	Tempr string `json:"temp"`
}

type Humidity struct {
	Humd string `json:"hum"`
}

type Light struct {
	Lum string `json:"lum"`
}

type Sensor struct {
	// Dummy string `json:"Dummy"`
	Accelerometer_Data Accelerometer `json:"Accelerometer"`
	Gyroscope_Data Gyroscope `json:"Gyroscope"`
	Temperature_Data Temperature `json:"Temperature"`
	Humidity_Data Humidity `json:"Humidity"`
	Light_Data Light `json:"Light"`
}

type IoTDataPoint struct {
	Device_ID string
	Timestamp string
	Sensor_Data Sensor `json:"Sensor"`
}

var IoTDataArray []IoTDataPoint // To be saved as gob file
var IoTDataElement IoTDataPoint // To be saved as gob file (new version)

/////

type Block struct { // An element of Blockchain
	Index             int
	Timestamp         string
	IoTDataPointEntry IoTDataPoint
	PrevHash          string
	ThisHash          string
}

type BlockMetaData struct {
	Index		int
	ThisHash 	string
}

var Blockchain []Block

///// LIST OF FUNCTIONS

func init() {

	gob.Register(IoTDataPoint{})
	gob.Register(Block{})
	gob.Register(map[string]interface{}{})

	log.SetFlags(log.Lshortfile)

	log.Printf("Welcome to DataKrew ProvDAT IoT Dashboard Server!")
	listenPort = flag.Int("port", 8085, "mux server listen port")
	dataDir = flag.String("dataDir", "data", "pathname of data directory to save IoT Data")
	flag.Parse()

	LoadIoTData() // load from existing files, if any

	StartTime = time.Now()
	StartTime = StartTime.AddDate(0, -6, 10) // random negative offset

	/////

	genesisBlock := Block{
		Index:     0,
		Timestamp: StartTime.Add(time.Duration(genRandInt(30000, 0)) * time.Second).Format("02-01-2006 15:04:05 Mon"),
		PrevHash:  "NULL",
	}
	genesisBlock.ThisHash = "GENESIS-BLOCK"; // calculateHash(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	/////
}

func main() {
	log.Fatal(launchMUXServer())
}

func launchMUXServer() error { // launch MUX server
	mux := makeMUXRouter()
	// log.Println("HTTP Server Listening on port:", *listenPort) // listenPort is a global flag
	log.Println("HTTP MUX server listening on " + GetMyIP() + ":" + os.Getenv("PORT")) // listenPort is a global const
	s := &http.Server{
		// Addr:           ":" + strconv.Itoa(*listenPort),
		Addr:           ":" + os.Getenv("PORT"),
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
	// muxRouter.HandleFunc("/post/{Temperature}/{Humidity}/{Sound}/{Gas}/{PIR}", handlePost_old).Methods("GET")
	muxRouter.HandleFunc("/post", handlePost_new).Methods("POST")
	muxRouter.HandleFunc("/blockchain", handleBlockchain).Methods("GET")
	muxRouter.HandleFunc("/blocklist", handleBlockList).Methods("GET")
	muxRouter.HandleFunc("/blockinfo/{Index}", handleBlockInfo).Methods("GET")
	return muxRouter
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHome() API called")
	io.WriteString(w, "You have entered the restricted zone. Trespassing is strictly prohibited. Defaulters will be reported.")
}

// func handlePost_old(w http.ResponseWriter, r *http.Request) {

// 	params := mux.Vars(r)

// 	newPoint := IoTDataPoint{}
// 	newPoint.SerialNo = len(IoTDataArray) + 1
// 	newPoint.Temperature = params["Temperature"]
// 	newPoint.Humidity = params["Humidity"]
// 	newPoint.Sound = params["Sound"]
// 	newPoint.Gas = params["Gas"]
// 	newPoint.PIR = params["PIR"]

// 	fmt.Println("Adding to SensorData:", newPoint)
// 	IoTDataArray = append(IoTDataArray, newPoint)
// 	gobCheck(writeIoTGob(IoTDataArray, len(IoTDataArray)))
// 	respondWithJSON(w, r, http.StatusCreated, newPoint)
// }

func handlePost_new(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var newPoint IoTDataPoint

    buf := new(bytes.Buffer)
    buf.ReadFrom(r.Body)
    newStr := buf.String()
    
    fmt.Println(newStr)
    decoder := json.NewDecoder(strings.NewReader(newStr))
    if err := decoder.Decode(&newPoint); err != nil {
        respondWithJSON(w, r, http.StatusBadRequest, r.Body)
        panic(err)
        return
    }
    defer r.Body.Close()

	fmt.Println("Adding to SensorData:", newPoint)
	IoTDataArray = append(IoTDataArray, newPoint)
	// gobCheck(writeIoTGob(IoTDataArray, len(IoTDataArray)))
	gobCheck(writeIoTGob(newPoint, len(IoTDataArray)))
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
		// gobCheck(readGob(&IoTDataArray, mostRecentFile))
		gobCheck(readGob(&IoTDataElement, mostRecentFile))
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

func handleBlockchain(w http.ResponseWriter, r *http.Request) {

	PrepareBlockchain()
	respondWithJSON(w, r, http.StatusCreated, Blockchain)
}

func handleBlockList(w http.ResponseWriter, r *http.Request) {
	
	PrepareBlockchain()
	var BlockList []BlockMetaData

	for j := 0 ; j < len(Blockchain) ; j++ {
		bmd := BlockMetaData{
			Index: Blockchain[j].Index,
			ThisHash: Blockchain[j].ThisHash,
		}
		BlockList = append (BlockList, bmd)
	}

	respondWithJSON(w, r, http.StatusCreated, BlockList)
}

func handleBlockInfo(w http.ResponseWriter, r *http.Request) {

	PrepareBlockchain()

	params := mux.Vars(r)
	InfoIndex, _ := strconv.Atoi(params["Index"])
	InfoBlock := Blockchain[InfoIndex]

	respondWithJSON(w, r, http.StatusCreated, InfoBlock)
}

func PrepareBlockchain() {
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

	log.Println("len(Blockchain) = ", len(Blockchain))
	log.Println("mostRecectFileNo = ", mostRecectFileNo)

	if (len(Blockchain) - 1 < mostRecectFileNo) {

		var TempIoTDataElement IoTDataPoint

		for i := len(Blockchain); i <= mostRecectFileNo; i++ {
			readFilePath := *dataDir + "/IoT-data-" + strconv.Itoa(i) + ".gob"
			gobCheck(readGob(&TempIoTDataElement, readFilePath))
			b := Block{
				Index:             i,
				Timestamp:         StartTime.AddDate(0, 0, genRandInt(3, 1)+(3*i)).Add(time.Duration(genRandInt(30000, 0)) * time.Second).Format("02-01-2006 15:04:05 Mon"), // random date increment
				IoTDataPointEntry: TempIoTDataElement,
				PrevHash:          Blockchain[len(Blockchain)-1].ThisHash,
			}
			b.ThisHash = calculateHash(b)
			Blockchain = append(Blockchain, b)
		}

		log.Println("len(Blockchain) = ", len(Blockchain))
	}
}

// SHA256 hashing
func calculateHash(b Block) string {
	record := strconv.Itoa(b.Index) + b.Timestamp + spew.Sdump(b.IoTDataPointEntry) + b.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func genRandString(n int) string { // generate Random String of length 'n'
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	val := make([]rune, n)
	for i := range val {
		myRandSource := rand.NewSource(time.Now().UnixNano())
		myRand := rand.New(myRandSource)
		val[i] = letterRunes[myRand.Intn(len(letterRunes))]
	}
	return string(val)
}

func genRandInt(n int, offset int) int { // generate Random Integer less than 'n' with an offset
	myRandSource := rand.NewSource(time.Now().UnixNano())
	myRand := rand.New(myRandSource)
	val := myRand.Intn(n) + offset
	return val
}

func GetMyIP() string {
	var MyIP string

	conn, err := gonet.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	} else {
		localAddr := conn.LocalAddr().(*gonet.UDPAddr)
		MyIP = localAddr.IP.String()
	}
	return MyIP
}
