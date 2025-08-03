package problemdetails

// Logger describes Error and Info functions.
type Logger interface {
	Error(args ...interface{})
}
