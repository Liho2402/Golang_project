package routes

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/gorilla/mux"
    "productservice/models"
)

var products = make(map[string]models.Product)

func SetupRouter() *mux.Router {
    r := mux.NewRouter()

    r.HandleFunc("/products", createProduct).Methods("POST")
    r.HandleFunc("/products", listProducts).Methods("GET")
    r.HandleFunc("/upload", uploadImage).Methods("POST")

    return r
}

func createProduct(w http.ResponseWriter, r *http.Request) {
    var p models.Product
    json.NewDecoder(r.Body).Decode(&p)
    products[p.ID] = p
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(p)
}

func listProducts(w http.ResponseWriter, r *http.Request) {
    category := r.URL.Query().Get("category")
    name := r.URL.Query().Get("name")
    minPriceStr := r.URL.Query().Get("min")
    maxPriceStr := r.URL.Query().Get("max")
    pageStr := r.URL.Query().Get("page")
    limitStr := r.URL.Query().Get("limit")

    var filtered []models.Product
    for _, p := range products {
        if category != "" && !strings.EqualFold(p.Category, category) {
            continue
        }
        if name != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) {
            continue
        }
        if minPriceStr != "" {
            min, _ := strconv.ParseFloat(minPriceStr, 64)
            if p.Price < min {
                continue
            }
        }
        if maxPriceStr != "" {
            max, _ := strconv.ParseFloat(maxPriceStr, 64)
            if p.Price > max {
                continue
            }
        }
        filtered = append(filtered, p)
    }

    // Пагинация
    page, _ := strconv.Atoi(pageStr)
    limit, _ := strconv.Atoi(limitStr)
    if page < 1 {
        page = 1
    }
    if limit < 1 {
        limit = 5
    }

    start := (page - 1) * limit
    end := start + limit
    if start > len(filtered) {
        start = len(filtered)
    }
    if end > len(filtered) {
        end = len(filtered)
    }

    paginated := filtered[start:end]
    json.NewEncoder(w).Encode(paginated)
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
    file, handler, err := r.FormFile("image")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    os.MkdirAll("uploads", os.ModePerm)
    filePath := filepath.Join("uploads", handler.Filename)
    dst, err := os.Create(filePath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    _, err = dst.ReadFrom(file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    imageURL := fmt.Sprintf("/uploads/%s", handler.Filename)
    json.NewEncoder(w).Encode(map[string]string{"url": imageURL})
}
