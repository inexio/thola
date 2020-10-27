//go:generate go run core/vfs/generate/filesystem_generate.go

package main

import "github.com/inexio/thola/cmd"

func main() {
	cmd.Execute()
}
