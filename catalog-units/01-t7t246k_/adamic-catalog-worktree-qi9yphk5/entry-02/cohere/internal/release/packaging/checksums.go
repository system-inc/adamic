package release

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ChecksumsFileName is the checksum list a release ships, in the dispatcher package and beside the staged
// packages, which is the copy attached to the GitHub release.
//
// The name and the line format are `shasum -a 256`'s, so the file is checked with a command every machine
// already has rather than with a tool of ours. A checker we wrote would be one more thing a reader has to
// trust before trusting the binary, which is the trust this file exists to remove.
const ChecksumsFileName = "SHA256SUMS"

// Checksums hashes every file the named platform packages ship under bin/, as SHA256SUMS lines.
//
// Each line is `<hex>  <package directory>/bin/<file>`, with forward slashes and sorted, so the file reads
// the same from every build machine and checks with `shasum -a 256 -c` from the directory holding the
// packages. That directory is the staging output, and it is also node_modules/@system-inc once npm has
// installed them, because the staging directories are named for the packages without their scope.
//
// Every file in bin/, not a list of the binaries we meant to ship. The platform manifest publishes all of
// bin/, so a file that lands there unlisted would reach every install unchecked. Hashed after signing,
// because codesign rewrites the Mach-O: a sum taken before it would refuse every signed binary.
func Checksums(outputDirectory string, packageDirectoryNames []string) ([]byte, error) {
	var lines []string
	for _, packageDirectoryName := range packageDirectoryNames {
		binDirectory := filepath.Join(outputDirectory, packageDirectoryName, "bin")
		entries, err := os.ReadDir(binDirectory)
		if err != nil {
			return nil, fmt.Errorf("reading %s to checksum it: %w", binDirectory, err)
		}

		hashed := 0
		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				// A directory or a link in bin/ is nothing the build writes, and hashing what a link points
				// at would vouch for a file outside the package.
				return nil, fmt.Errorf("%s holds %s, which is not a regular file, so its checksum would not describe what installs", binDirectory, entry.Name())
			}
			digest, err := sha256File(filepath.Join(binDirectory, entry.Name()))
			if err != nil {
				return nil, err
			}
			lines = append(lines, digest+"  "+packageDirectoryName+"/bin/"+entry.Name())
			hashed++
		}
		if hashed == 0 {
			return nil, fmt.Errorf("%s holds no files, so there is no binary to checksum", binDirectory)
		}
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no platform packages to checksum")
	}
	slices.SortFunc(lines, func(left string, right string) int {
		// By path, the part after the hash and its two spaces.
		return strings.Compare(left[sha256.Size*2+2:], right[sha256.Size*2+2:])
	})
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

// sha256File is a file's SHA-256 in lowercase hex, the spelling `shasum` prints and compares.
func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening %s to checksum it: %w", path, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("reading %s to checksum it: %w", path, err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
