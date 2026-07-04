package replace

import (
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

type Replace struct {
	Package   string
	LocalPath string
}

func Parse(file string) ([]Replace, error) {
	modBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read mod file: %w", err)
	}

	mod, err := modfile.Parse(file, modBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mod file: %w", err)
	}

	var replaces []Replace
	for _, replace := range mod.Replace {
		replaces = append(replaces, Replace{
			Package:   replace.Old.Path,
			LocalPath: replace.New.Path,
		})
	}

	return replaces, nil
}
