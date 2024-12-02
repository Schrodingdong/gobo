package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
  indexContent, err := os.ReadFile("./html/index.html")
  host := "localhost"
  port := "9999"

  if err != nil { panic(err) }
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    log.Println("Received !!")
    w.Write(indexContent)
  })
  
  log.Printf("Hosting on: %s:%s\n", host, port)
  panic(http.ListenAndServe(host+":"+port, nil))
}
