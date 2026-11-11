package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	sources := []string{"../foo.json", "foo/../../bar.json", "C:\\..\\windows\\system32", "my..file.json", "/foo/../bar.json"}
	for _, source := range sources {
		cleaned := filepath.Clean(source)
		hasPrefix := strings.HasPrefix(cleaned, "..")
		fmt.Printf("%q -> %q -> %v\n", source, cleaned, hasPrefix)
	}
}
