package lambdas

// MapMessage represent a value to emit
type MapMessage struct {
	Key      string  `json:"key,omitempty"`
	Value    float64 `json:"value,omitempty"`
	Type     float64 `json:"type,omitempty"`
	EmptyVal bool    `json:"empty,omitempty"`
}
