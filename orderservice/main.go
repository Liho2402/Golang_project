package main

import (
    "log"
    "net/http"
    "orderservice/routes"
)

func main() {
    router := routes.SetupRouter()
    log.Println("OrderService started on :8081")
    log.Fatal(http.ListenAndServe(":8081", router))
}
