package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/schrodi/gobo/algorithm"
)

var srcAddr = "localhost:8080"
var destAddr = ""
var clientBucketMap = make(map[string]*algorithm.Bucket)
var maxBucketSize = 5
var refilDelay = 15 //s
var proxyClient *http.Client
func init() {
  args := os.Args
  if len(args) == 1 {
    printBanner()
    printUsage()
    os.Exit(0)
  } else {
    for i := 1; i < len(args); i ++ {
      if args[i] == "--src" && len(args) > i+1 {
        srcAddr = args[i+1]
        i = i + 1
      } else if args[i] == "--dest" && len(args) > i+1 {
        destAddr = args[i+1]
        i = i + 1
      } else if args[i] == "--max-bucket-size" && len(args) > i+1 {
        mbs, err := strconv.Atoi(args[i+1]) 
        if err != nil { panic(err) }
        maxBucketSize = mbs
        i = i + 1
      } else if args[i] == "--refil-delay" && len(args) > i+1{
        rd , err := strconv.Atoi(args[i+1]) 
        if err != nil { panic(err) }
        refilDelay = rd
        i = i + 1
      }
    }
  }
  if destAddr == "" {
    fmt.Println("Needs --dest flag")
    printUsage()
    os.Exit(1)
  }
  // Formatting
  srcAddr  = formatAddress(srcAddr, false)
  destAddr = formatAddress(destAddr, true)
  log.Printf("Src address %q", srcAddr)
  log.Printf("Dest address %q", destAddr)
  
  // Start refil routine
  go algorithm.RefillRoutine(&clientBucketMap)

  // Init proxy client
  proxyClient = &http.Client{}
}


func createBucket(clientAddr string) *algorithm.Bucket{
  b := &algorithm.Bucket{
    Size: maxBucketSize,
    RefillDelay: time.Duration(refilDelay) * time.Second,
    BucketFill: maxBucketSize - 1,
  }
  clientBucketMap[clientAddr] = b
  return b
}

func proxyRequest(r *http.Request) (*http.Response, error) {
  proxyReq, err := http.NewRequest(
    r.Method,
    destAddr,
    r.Body,
  )
  if err != nil { return nil, err }
  proxyResp, err := proxyClient.Do(proxyReq)
  if err != nil { return nil, err }
  return proxyResp, nil
}

func limitAndHandle(w http.ResponseWriter, r *http.Request) {
    clientAddr := r.RemoteAddr
    b, bExists := clientBucketMap[clientAddr]
    if !bExists {
      log.Printf("Creating bucket for %q ...\n", clientAddr)
      b = createBucket(clientAddr)
    } else {
      tokensLeft := b.ConsumeToken()
      log.Printf("Tokens left for %q: %v\n", clientAddr, tokensLeft)
      if tokensLeft > 0 {
        proxyResp, err := proxyRequest(r)
        if err != nil { panic(err) }
        w = initResponseWriterFromResponse(w, proxyResp)
      } else {
        w.Write([]byte("429 - TooManyRequests :/"))
      }
    }
  }


func main() {
  http.HandleFunc("/", limitAndHandle)

  log.Println("Listening on port 8080...")
  panic(http.ListenAndServe(srcAddr, nil))
}
