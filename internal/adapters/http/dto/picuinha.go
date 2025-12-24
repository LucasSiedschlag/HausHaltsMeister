package dto

type CreatePersonRequest struct {
	Name  string `json:"name"`
	Notes string `json:"notes"`
}

type PersonResponse struct {
	ID      int32   `json:"id"`
	Name    string  `json:"name"`
	Notes   string  `json:"notes"`
	Balance float64 `json:"balance"`
}

type AddEntryRequest struct {
	PersonID       int32   `json:"person_id"`
	Kind           string  `json:"kind"` // PLUS or MINUS
	Amount         float64 `json:"amount"`
	CashFlowID     *int32  `json:"cash_flow_id"` // Optional
	AutoCreateFlow bool    `json:"auto_create_flow"`
}

type PicuinhaEntryResponse struct {
	ID         int32   `json:"id"`
	PersonID   int32   `json:"person_id"`
	Amount     float64 `json:"amount"`
	Kind       string  `json:"kind"`
	CashFlowID *int32  `json:"cash_flow_id"`
	CreatedAt  string  `json:"created_at"`
}
