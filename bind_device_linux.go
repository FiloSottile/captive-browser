package main

import (
	"log"
	"syscall"
)

func bindToDevice(device string) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {
			err := syscall.BindToDevice(int(fd), device)
			if err != nil {
				log.Fatalf("Failed to bind to %q: %s", device, err)
			}
		})
	}
}
