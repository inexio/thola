package main

import "github.com/inexio/thola/cmd"

// 'go generate' generates the mocks needed for the tests. This requires mockery as a dependency.
// Run 'go generate ./...' in the root folder of the project to generate all mocks.
//go:generate go get -d github.com/vektra/mockery/v2@v2.26.1

func main() {
	cmd.Execute()
}
