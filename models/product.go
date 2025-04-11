package models

type Product struct {
    ID       string  `json:"id"`
    Name     string  `json:"name"`
    Category string  `json:"category"`
    Price    float64 `json:"price"`
    ImageURL string  `json:"image_url"`
}

