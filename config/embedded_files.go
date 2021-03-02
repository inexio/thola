package config

import (
	"embed"
)

//go:embed device-classes mappings
var FileSystem embed.FS
