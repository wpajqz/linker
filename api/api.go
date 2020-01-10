package api

type API interface {
	Dial(network, address string) error
}
