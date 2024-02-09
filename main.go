package main

import (
  "fmt"

  api "main/api"
)

func main() {
  fmt.Println("Sup Buddy")
  // Create a new server object.
  server := api.NewApiServer(":8080")
  //Run the server on the port 8080
  server.Run()
}

