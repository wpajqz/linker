package api

type API interface {
	Run(debug bool) error
}
