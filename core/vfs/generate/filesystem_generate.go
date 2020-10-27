// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(http.Dir("core/config"), vfsgen.Options{
		PackageName:  "vfs",
		VariableName: "FileSystem",
		Filename:     "core/vfs/filesystem-vfsdata.go",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
