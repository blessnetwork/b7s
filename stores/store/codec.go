package store

type Codec interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}
