package pkg

type Price struct {
	ID        int     `json:"id"`
	CreatedAt string  `json:"created_at"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Price     float64 `json:"price"`
}
