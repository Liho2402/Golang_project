package routes

import (
    "encoding/json"
    "net/http"
    "orderservice/models"

    "github.com/gorilla/mux"
)

var orders = make(map[string]models.Order)

func SetupRouter() *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/orders", createOrder).Methods("POST")
    r.HandleFunc("/orders/{id}", getOrder).Methods("GET")
    r.HandleFunc("/orders/{id}/status", updateStatus).Methods("PUT")
    r.HandleFunc("/payments", mockPayment).Methods("POST")
    return r
}

func createOrder(w http.ResponseWriter, r *http.Request) {
    var o models.Order
    json.NewDecoder(r.Body).Decode(&o)
    orders[o.ID] = o
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(o)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    order, exists := orders[id]
    if !exists {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(order)
}

func updateStatus(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    var update struct {
        Status string `json:"status"`
    }
    json.NewDecoder(r.Body).Decode(&update)
    order := orders[id]
    order.Status = update.Status
    orders[id] = order
    json.NewEncoder(w).Encode(order)
}

func mockPayment(w http.ResponseWriter, r *http.Request) {
    var req struct {
        OrderID string `json:"order_id"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    order := orders[req.OrderID]
    order.Status = "paid"
    orders[req.OrderID] = order
    json.NewEncoder(w).Encode(order)
}
