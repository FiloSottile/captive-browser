// +build !linux

package main

import (
	"fmt"
	"runtime"
	"syscall"
)

func bindToDevice(device string) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		return fmt.Errorf("bind-device not supported on %s", runtime.GOOS)
	}
}
