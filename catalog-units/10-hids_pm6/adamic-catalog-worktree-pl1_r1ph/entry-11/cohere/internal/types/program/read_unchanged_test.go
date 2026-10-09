package program

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

// encodeUtf16LittleEndian is text as UTF-16 LE behind its byte order mark, the bytes a UTF-16 editor saves.
func encodeUtf16LittleEndian(text string) []byte {
	encoded := []byte{0xFF, 0xFE}
	for _, unit := range utf16.Encode([]rune(text)) {
		encoded = append(encoded, byte(unit), byte(unit>>8))
	}
	return encoded
}

// The program's copy of a file is its bytes only where decodeBytes left them alone (#hkv1hgp). ReadUnchanged
// says yes for a plain file, and no for a file behind a UTF-8 byte order mark, a UTF-16 one, and a UTF-16 one
// whose decoded length is exactly its size on disk, which only the look at its first bytes can turn away. Both
// with a content pack and without one, since each records the stat on its own path.
func TestTheProgramsCopyIsTheFilesBytesOnlyWithoutAByteOrderMark(t *testing.T) {
	t.Parallel()
	// Two more three-byte characters than ASCII bytes makes UTF-16's 2 + 2 units equal UTF-8's length.
	ascii := "export const sameLength = '';\n"
	sameLength := ascii[:len(ascii)-4] + strings.Repeat("中", len(ascii)+2) + ascii[len(ascii)-4:]
	if size, decoded := len(encodeUtf16LittleEndian(sameLength)), len(sameLength); size != decoded {
		t.Fatalf("the same-length fixture is %d bytes on disk and %d decoded, so it tests nothing", size, decoded)
	}
	files := map[string]struct {
		bytes  []byte
		usable bool
	}{
		"Plain.ts":      {[]byte("export const plain = 1;\n"), true},
		"Utf8Mark.ts":   {append([]byte{0xEF, 0xBB, 0xBF}, "export const marked = 1;\n"...), false},
		"Utf16.ts":      {encodeUtf16LittleEndian("export const wide = 1;\n"), false},
		"SameLength.ts": {encodeUtf16LittleEndian(sameLength), false},
	}
	for _, withPack := range []bool{true, false} {
		t.Run(map[bool]string{true: "with a content pack", false: "without one"}[withPack], func(t *testing.T) {
			t.Parallel()
			directory, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			for name, file := range files {
				if err := os.WriteFile(filepath.Join(directory, name), file.bytes, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			configuration := `{"compilerOptions":{"strict":true,"module":"esnext","target":"esnext"},"include":["*.ts"]}`
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
				t.Fatal(err)
			}
			var pack *ContentPack
			if withPack {
				if pack, err = OpenContentPack(t.TempDir()); err != nil {
					t.Fatal(err)
				}
			}
			graph, err := Build(Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, ContentPack: pack})
			if err != nil {
				t.Fatal(err)
			}
			// The plain file first: a guard that refused everything would pass every case below.
			for _, name := range []string{"Plain.ts", "Utf8Mark.ts", "Utf16.ts", "SameLength.ts"} {
				path := filepath.Join(directory, name)
				sourceFile := graph.Program.GetSourceFile(tspath.RootedFilePath(path))
				if sourceFile == nil {
					t.Fatalf("%s is not in the program", name)
				}
				if got := graph.ReadUnchanged(path, sourceFile.Text()); got != files[name].usable {
					t.Errorf("%s: ReadUnchanged is %v, want %v", name, got, files[name].usable)
				}
			}
		})
	}
}
