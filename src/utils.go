package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

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
