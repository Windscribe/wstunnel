package cli

import "runtime/debug"

const (
	// BufferSize is the size of the intermediate buffer for network packets
	BufferSize = 1024
)

// Runner defines a basic interface with only a Run() function
type Runner interface {
	Run() error
}

// goSafely runs fn in a new goroutine. A panic in it is logged and stops only that goroutine: an unrecovered panic
// would end the whole host process, which on iOS and Android is the app's VPN tunnel.
func goSafely(name string, fn func()) {
	go func() {
		defer recoverPanic(name)
		fn()
	}()
}

// recoverPanic logs a panic in progress, with its stack, and stops it. Call it with defer.
func recoverPanic(name string) {
	if r := recover(); r != nil {
		Logger.Errorf("Recovered from a panic in %s: %v\n%s", name, r, debug.Stack())
	}
}
