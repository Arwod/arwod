package scriptengine

import (
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/buffer"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/process"
	"github.com/dop251/goja_nodejs/require"
)

// VMPoolItem represents a single VM instance in the pool
type VMPoolItem struct {
	mux  sync.Mutex
	busy bool
	vm   *goja.Runtime
}

// VMPool manages a pool of JavaScript VM instances
type VMPool struct {
	mux     sync.RWMutex
	factory func() *goja.Runtime
	items   []*VMPoolItem
}

// NewVMPool creates a new VM pool with pre-warmed VMs
func NewVMPool(size int, factory func() *goja.Runtime) *VMPool {
	pool := &VMPool{
		factory: factory,
		items:   make([]*VMPoolItem, size),
	}

	for i := 0; i < size; i++ {
		vm := pool.factory()
		pool.items[i] = &VMPoolItem{vm: vm}
	}

	return pool
}

// Run executes a function with a VM from the pool
func (p *VMPool) Run(call func(vm *goja.Runtime) error) error {
	p.mux.RLock()

	// Try to find a free item
	var freeItem *VMPoolItem
	for _, item := range p.items {
		item.mux.Lock()
		if item.busy {
			item.mux.Unlock()
			continue
		}
		item.busy = true
		item.mux.Unlock()
		freeItem = item
		break
	}

	p.mux.RUnlock()

	// Create a new one-off item if all pool items are busy
	if freeItem == nil {
		return call(p.factory())
	}

	execErr := call(freeItem.vm)

	// Free the VM
	freeItem.mux.Lock()
	freeItem.busy = false
	freeItem.mux.Unlock()

	return execErr
}

// CreateBaseVM creates a basic JavaScript VM with Node.js modules
func CreateBaseVM() *goja.Runtime {
	vm := goja.New()

	// Enable Node.js modules
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	buffer.Enable(vm)
	process.Enable(vm)

	// Set basic globals
	vm.Set("setTimeout", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}

		callback := call.Arguments[0]
		delay := call.Arguments[1].ToInteger()

		go func() {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			if fn, ok := goja.AssertFunction(callback); ok {
				fn(goja.Undefined())
			}
		}()

		return goja.Undefined()
	})

	vm.Set("setInterval", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}

		callback := call.Arguments[0]
		interval := call.Arguments[1].ToInteger()

		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()

			for range ticker.C {
				if fn, ok := goja.AssertFunction(callback); ok {
					fn(goja.Undefined())
				}
			}
		}()

		return goja.Undefined()
	})

	return vm
}