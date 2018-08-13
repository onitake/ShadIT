
package main

// Gpio is the abstraction of a single GPIO line.
// For each supported architecture, a specialised implementation is provided.
type Gpio interface {
	// Init sets up the GPIO line for input or output, according to
	// its platform specific implementation.
	Init() error
	// Set changes the digital state of this GPIO line, either logical 1
	// (high) or 0 (low).
	// The exact outcome is machine- and platform-defined.
	// May cause unexpected results if the GPIO line is configured for input.
	Set(state bool) error
	// Get obtains the current logical state of the GPIO line.
	// The exact result is machine- and platform-defined.
	Get() (bool, error)
}
