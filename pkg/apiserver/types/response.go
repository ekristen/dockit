package types

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []error     `json:"errors,omitempty"`
}
