package mapper

// MapInt represent a map[string]int data type input
type MapInt struct {
	Key      string `json:"key,omitempty"`
	Value    int    `json:"value,omitempty"`
	EmptyVal bool   `json:"empty,omitempty"`
}
