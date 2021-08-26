package main

import "github.com/inexio/thola/cmd"

// go generate generates the mocks needed for tests. this requires mockery as a dependency
//go:generate go get github.com/vektra/mockery/v2

func main() {
	cmd.Execute()
}
