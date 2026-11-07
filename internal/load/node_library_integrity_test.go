package load

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// The archives are the published npm tarballs, not another editable declaration
// copy. Their integrity values are recorded from the locked registry packages.
func publishedNodeTypeFiles(test *testing.T, archive, integrity string) map[string][]byte {
	test.Helper()
	data, readError := os.ReadFile(archive)
	if readError != nil {
		test.Fatal(readError)
	}
	digest := sha512.Sum512(data)
	if got := "sha512-" + base64.StdEncoding.EncodeToString(digest[:]); got != integrity {
		test.Fatalf("published package integrity: %s != %s", got, integrity)
	}
	compressed, gzipError := gzip.NewReader(bytes.NewReader(data))
	if gzipError != nil {
		test.Fatal(gzipError)
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	files := map[string][]byte{}
	for {
		entry, nextError := reader.Next()
		if nextError == io.EOF {
			break
		}
		if nextError != nil {
			test.Fatal(nextError)
		}
		if entry.Typeflag != tar.TypeReg {
			continue
		}
		_, name, ok := strings.Cut(entry.Name, "/")
		if !ok || !fs.ValidPath(name) {
			test.Fatalf("invalid published file name %q", entry.Name)
		}
		data, readError := io.ReadAll(reader)
		if readError != nil {
			test.Fatal(readError)
		}
		files[name] = data
	}
	return files
}

func TestEmbeddedNodeTypesIntegrity(test *testing.T) {
	test.Parallel()
	data, readError := os.ReadFile("node_types_manifest.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var manifest struct {
		Packages []struct {
			Name, Version, URL, Integrity, Directory string
			Files                                    map[string]string
		}
	}
	if decodeError := json.Unmarshal(data, &manifest); decodeError != nil {
		test.Fatal(decodeError)
	}
	expected := map[string]string{
		"@types/node@25.3.3":  "sha512-DpzbrH7wIcBaJibpKo9nnSQL0MTRdnWttGyE5haGwK86xgMOkFLp7vEyfQPGLOJh5wNYiJ3V9PmUMDhV9u8kkQ==",
		"undici-types@7.18.2": "sha512-AsuCzffGHJybSaRrmr5eHr81mwJU3kjw6M+uprWvCXiNeN9SOGwQ3Jn8jb8m3Z6izVgknn1R0FTCEAP2QrLY/w==",
	}
	if len(manifest.Packages) != len(expected) {
		test.Fatal("embedded package manifest has the wrong package set")
	}
	for _, pin := range manifest.Packages {
		integrity, ok := expected[pin.Name+"@"+pin.Version]
		if !ok || integrity != pin.Integrity {
			test.Fatalf("unrecognized embedded package pin %s@%s", pin.Name, pin.Version)
		}
		name := path.Base(pin.Name) + "-" + pin.Version + ".tgz"
		published := publishedNodeTypeFiles(test, filepath.Join("testdata", "node_types", name), integrity)
		if len(published) != len(pin.Files) {
			test.Fatal("published file list differs from hash manifest")
		}
		for name, truth := range published {
			want := sha256.Sum256(truth)
			if pin.Files[name] != hex.EncodeToString(want[:]) {
				test.Fatalf("recorded hash differs from published package: %s/%s", pin.Name, name)
			}
			file := "node_types/" + pin.Directory + "/" + name
			embedded, readFileError := embeddedNodeTypes.ReadFile(file)
			if readFileError != nil {
				test.Fatal(readFileError)
			}
			got := sha256.Sum256(embedded)
			if got != want {
				test.Errorf("embedded byte mismatch: %s/%s", pin.Name, name)
			}
		}
		walkDirectoryError := fs.WalkDir(embeddedNodeTypes, "node_types/"+
			pin.Directory, func(file string, entry fs.DirEntry, pathError error) error {
			if pathError != nil {
				return pathError
			}
			if entry.IsDir() {
				return nil
			}
			name := strings.TrimPrefix(file, "node_types/"+pin.Directory+"/")
			if _, ok := published[name]; !ok {
				test.Errorf("unpublished embedded file: %s", file)
			}
			return nil
		})
		if walkDirectoryError != nil {
			test.Fatal(walkDirectoryError)
		}
	}
}
