package testdata

import _ "embed"

//go:embed valid-go.mod
var GoMod []byte

//go:embed replaces-go.mod
var GoModWithReplaces []byte

//go:embed valid-config.yaml
var ConfigYaml []byte
