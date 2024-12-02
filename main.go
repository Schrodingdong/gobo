package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

// Util Functions
func formatAddress(addr string, withProtocol bool) string{
  rg := regexp.MustCompile(`((.+):\/\/)*(.+)*:(\d+)`)
  match := rg.FindStringSubmatch(addr)
  protocol := match[2]
  host := match[3]
  if protocol == "" {
    protocol = "http"
  }
  if host == "" {
    host = "localhost"
  }
  port := match[4]
  if withProtocol {
    return protocol + "://" + host + ":" + port
  } else {
    return host + ":" + port
  }
}

func initResponseWriterFromResponse(w http.ResponseWriter, res *http.Response) http.ResponseWriter{
  resBody, err := io.ReadAll(res.Body)
  if err != nil { panic(err) }
  w.Write(resBody)
  for k := range res.Header {
    headerVal := res.Header.Get(k)
    w.Header().Add(k, headerVal)
  }
  return w
}

func printBanner() {
  fmt.Printf(` _____     _____
|   __|___| __  |___ 
|  |  | . | __ -| . |
|_____|___|_____|___|

GoBo (Go One By One) - Rate limiter written in go !

`)
}

func printUsage() {
  fmt.Println(`Usage
  gobo [--src ip:port] [--max-bucket-size n] [--refil-delay n] --dest ip:port
Example:
  gobo --dest :8080
  gobo --dest localhost:8080
  gobo --dest 127.0.0.1:8888`)
}


// Token Bucket algorithm
type Bucket struct {
  Size int                    // Num of tokens
  RefillDelay time.Duration   // Timestamped
  lastRefill time.Time
  bucketFill int              // Available tokens
}

func (b *Bucket) refill() {
  b.bucketFill = b.Size
}

func (b *Bucket) consumeToken() int {
  if b.bucketFill > 0 {
    b.bucketFill -= 1
  }
  return b.bucketFill
}

func refillRoutine(bl *map[string]*Bucket) {
  for {
    t := time.Now()
    for k, b := range *bl {
      if b.RefillDelay < t.Sub(b.lastRefill) && b.Size > b.bucketFill{
        log.Printf("Refilling bucket for: %q\n", k)
        b.refill()
        b.lastRefill = t
      }
    } 
    time.Sleep(1 * time.Second)
  }
}


func main() {
  var srcAddr  string = "localhost:8080"
  var destAddr string
  var maxBucketSize  = 5
  var refilDelay = 15 //s
  var clientBucketMap = make(map[string]*Bucket)
  var proxyClient = &http.Client{}
  go refillRoutine(&clientBucketMap)

  args := os.Args
  if len(args) == 1 {
    printBanner()
    printUsage()
    return 
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
    return
  }
  // Formatting
  srcAddr  = formatAddress(srcAddr, false)
  destAddr = formatAddress(destAddr, true)
  log.Printf("Src address %q", srcAddr)
  log.Printf("Dest address %q", destAddr)
  // Load file content
  tooManyReqContent, err := os.ReadFile("./html/429.html")
  if err != nil { panic(err) }
  // Add handlers
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    clientAddr := r.RemoteAddr
    b, bExists := clientBucketMap[clientAddr]
    if !bExists {
      b = &Bucket{
        Size: maxBucketSize,
        RefillDelay: time.Duration(refilDelay) * time.Second,
        bucketFill: maxBucketSize - 1,
      }
      clientBucketMap[clientAddr] = b
      log.Printf("Creating bucket for %q ...\n", clientAddr)
    } else {
      tokensLeft := b.consumeToken()
      if tokensLeft > 0 {
        log.Printf("Consuming token for %q. Tokens left: %v\n", clientAddr, tokensLeft)
        // Proxy request
        proxyReq, err := http.NewRequest(
          r.Method,
          destAddr,
          r.Body,
        )
        if err != nil { panic(err) }
        proxyResp, err := proxyClient.Do(proxyReq)
        if err != nil { panic(err) }
        w = initResponseWriterFromResponse(w, proxyResp)
      } else {
        log.Printf("No more tokens left for %q :/\n", clientAddr)
        w.Write(tooManyReqContent)
      }
    }
  })
  // Start server
  log.Println("Listening on port 8080...")
  panic(http.ListenAndServe(srcAddr, nil))
}
