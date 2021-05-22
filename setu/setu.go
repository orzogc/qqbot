package setu

type Setu interface {
	GetImage() ([][]byte, error)
}
