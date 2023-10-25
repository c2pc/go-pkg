// Package config provides the Driver interface.
// All config drivers must implement this interface, register themselves,
// optionally provide a `WithInstance` function and pass the tests
// in package config/testing.
package config

import (
	"fmt"
	iurl "github.com/c2pc/go-pkg/config_migration/url"
	"io"
	"sync"
)

var (
	ErrLocked    = fmt.Errorf("can't acquire lock")
	ErrNotLocked = fmt.Errorf("can't unlock, as not currently locked")
)

const NilVersion int = -1

var driversMu sync.RWMutex
var drivers = make(map[string]Driver)

// Driver is the interface every config driver must implement.
type Driver interface {
	// Open returns a new driver instance configured with parameters
	// coming from the URL string. Migrate will call this function
	// only once per instance.
	Open(url string) (Driver, error)

	// Close closes the underlying config instance managed by the driver.
	// Migrate will call this function only once per instance.
	Close() error

	// Lock should acquire a config lock so that only one migration process
	// can run at a time. Migrate will call this function before Run is called.
	// If the implementation can't provide this functionality, return nil.
	// Return config.ErrLocked if config is already locked.
	Lock() error

	// Unlock should release the lock. Migrate will call this function after
	// all migrations have been run.
	Unlock() error

	// Run applies a migration to the config. migration is guaranteed to be not nil.
	Run(migration io.Reader) error

	// Version returns the currently active version and if the config is dirty.
	// When no migration has been applied, it must return version -1.
	// Dirty means, a previous migration failed and user interaction is required.
	Version() (version int, err error)

	// Drop deletes everything in the config.
	// Note that this is a breaking action, a new call to Open() is necessary to
	// ensure subsequent calls work as expected.
	Drop() error
}

// Open returns a new driver instance.
func Open(path string) (Driver, error) {
	scheme, err := iurl.SchemeFromURL(path)
	if err != nil {
		return nil, err
	}

	driversMu.RLock()
	d, ok := drivers[scheme]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("config driver: unknown driver %v (forgotten import?)", scheme)
	}

	return d.Open(path)
}

// Register globally registers a driver.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// List lists the registered drivers
func List() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	names := make([]string, 0, len(drivers))
	for n := range drivers {
		names = append(names, n)
	}
	return names
}
