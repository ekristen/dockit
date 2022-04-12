package types

import "encoding/json"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Errors  Errors      `json:"errors,omitempty"`
}

type Errors []error

func (e Errors) MarshalJSON() ([]byte, error) {
	res := make([]interface{}, len(e))
	for i, e := range e {
		if _, ok := e.(json.Marshaler); ok {
			res[i] = e // e knows how to marshal itself
		} else {
			res[i] = e.Error() // Fallback to the error string
		}
	}
	return json.Marshal(res)
}
