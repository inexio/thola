//go:generate go run core/vfs/generate/filesystem_generate.go

package main

import "thola/cmd"

func main() {
	cmd.Execute()
}
