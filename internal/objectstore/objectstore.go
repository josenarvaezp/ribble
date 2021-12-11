package objectstore

// Object represent a cloud object
type Object struct {
	Bucket string
	Key    string
	Size   int64
}

// ObjectRange represents an cloud object with its range specified
type ObjectRange struct {
	Object      Object
	InitialByte int64
	FinalByte   int64
}

// Bucket represents a cloud bucket
type Bucket struct {
	Name string `yaml:"bucket"`
	Keys []string
}

// NewObjectWithRange creates a new objectRange
func NewObjectWithRange(object Object, initialByte int64, finalByte int64) ObjectRange {
	return ObjectRange{
		Object:      object,
		InitialByte: initialByte,
		FinalByte:   finalByte,
	}
}
