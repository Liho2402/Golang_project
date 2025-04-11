package main

import (
    "log"
    "net/http"
    "productservice/routes"
)

func main() {
    r := routes.SetupRouter()
    log.Println("ProductService started on :8082")
    http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
    log.Fatal(http.ListenAndServe(":8082", r))
}
