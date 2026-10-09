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
func publishedNodeTypeFiles(t *testing.T, archive, integrity string) map[string][]byte {
	t.Helper()
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha512.Sum512(data)
	if got := "sha512-" + base64.StdEncoding.EncodeToString(digest[:]); got != integrity {
		t.Fatalf("published package integrity: %s != %s", got, integrity)
	}
	compressed, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	files := map[string][]byte{}
	for {
		entry, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if entry.Typeflag != tar.TypeReg {
			continue
		}
		_, name, ok := strings.Cut(entry.Name, "/")
		if !ok || !fs.ValidPath(name) {
			t.Fatalf("invalid published file name %q", entry.Name)
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		files[name] = data
	}
	return files
}

func TestEmbeddedNodeTypesIntegrity(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("node_types_manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Packages []struct {
			Name, Version, URL, Integrity, Directory string
			Files                                    map[string]string
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{
		"@types/node@25.3.3":  "sha512-DpzbrH7wIcBaJibpKo9nnSQL0MTRdnWttGyE5haGwK86xgMOkFLp7vEyfQPGLOJh5wNYiJ3V9PmUMDhV9u8kkQ==",
		"undici-types@7.18.2": "sha512-AsuCzffGHJybSaRrmr5eHr81mwJU3kjw6M+uprWvCXiNeN9SOGwQ3Jn8jb8m3Z6izVgknn1R0FTCEAP2QrLY/w==",
	}
	if len(manifest.Packages) != len(expected) {
		t.Fatal("embedded package manifest has the wrong package set")
	}
	for _, pin := range manifest.Packages {
		integrity, ok := expected[pin.Name+"@"+pin.Version]
		if !ok || integrity != pin.Integrity {
			t.Fatalf("unrecognized embedded package pin %s@%s", pin.Name, pin.Version)
		}
		name := path.Base(pin.Name) + "-" + pin.Version + ".tgz"
		published := publishedNodeTypeFiles(t, filepath.Join("testdata", "node_types", name), integrity)
		if len(published) != len(pin.Files) {
			t.Fatal("published file list differs from hash manifest")
		}
		for name, truth := range published {
			want := sha256.Sum256(truth)
			if pin.Files[name] != hex.EncodeToString(want[:]) {
				t.Fatalf("recorded hash differs from published package: %s/%s", pin.Name, name)
			}
			file := "node_types/" + pin.Directory + "/" + name
			embedded, err := embeddedNodeTypes.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			got := sha256.Sum256(embedded)
			if got != want {
				t.Errorf("embedded byte mismatch: %s/%s", pin.Name, name)
			}
		}
		err := fs.WalkDir(embeddedNodeTypes, "node_types/"+pin.Directory, func(file string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			name := strings.TrimPrefix(file, "node_types/"+pin.Directory+"/")
			if _, ok := published[name]; !ok {
				t.Errorf("unpublished embedded file: %s", file)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
