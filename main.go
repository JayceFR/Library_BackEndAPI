package main

import (
  "fmt"

  api "main/api"
)

func main() {
  fmt.Println("Sup Buddy")
  server := api.NewApiServer(":8080")
  server.Run()
}
