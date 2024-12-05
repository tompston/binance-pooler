package syro

// Standartized response of the API
type ApiResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Params  any    `json:"params,omitempty"`
}
