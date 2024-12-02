package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
  indexContent, err := os.ReadFile("./index.html")
  if err != nil { panic(err) }
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received !!")
    w.Write(indexContent)
  })

  panic(http.ListenAndServe(":9999", nil))
}
