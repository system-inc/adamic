package boundedrun

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"strings"
)

// Cache clients must include the executed helper implementation in their keys.
// A changed timeout/capture implementation can change an observation even when
// the caller's source and Node bytes are unchanged.
//
//go:embed *.go
var sources embed.FS

func Identity() string {
	entries, readDirectoryError := sources.ReadDir(".")
	if readDirectoryError != nil {
		panic(readDirectoryError)
	}
	hash := sha256.New()
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		data, readFileError := sources.ReadFile(entry.Name())
		if readFileError != nil {
			panic(readFileError)
		}
		fmt.Fprintf(hash, "%d:%s%d:", len(entry.Name()), entry.Name(), len(data))
		hash.Write(data)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
