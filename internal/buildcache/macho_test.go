package buildcache

import (
	"bytes"
	"context"
	"debug/macho"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Generate a real small Darwin executable on every host, then patch only bytes found with debug/macho.
func TestMachOAuditIgnoresOnlyUUIDAndSignature(t *testing.T) {
	t.Parallel()
	content := machOFixture(t)
	file, err := macho.NewFile(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	uuid, signature := -1, -1
	offset := 32
	for _, load := range file.Loads {
		raw := load.Raw()
		switch file.ByteOrder.Uint32(raw) {
		case 0x1b:
			uuid = offset + 8
		case 0x1d:
			if file.ByteOrder.Uint32(raw[12:16]) == 0 {
				t.Fatal("empty signature")
			}
			signature = int(file.ByteOrder.Uint32(raw[8:12]))
		}
		offset += len(raw)
	}
	text := file.Section("__text")
	if uuid < 0 || signature < 0 || text == nil || text.Size == 0 {
		t.Fatalf("fixture lacks UUID, signature or __text: %d, %d, %v", uuid, signature, text)
	}
	for name, patch := range map[string]int{"uuid": uuid, "signature": signature, "text": int(text.Offset)} {
		t.Run(name, func(t *testing.T) {
			changed := bytes.Clone(content)
			changed[patch] ^= 1
			fetched := t.TempDir()
			write(t, fetched, "product", string(changed))
			before := bytes.Clone(changed)
			err := audit(strings.Repeat("a", 64), "Mach-O", fetched, func(directory string) error {
				return os.WriteFile(filepath.Join(directory, "product"), content, 0o644)
			})
			if name == "text" {
				var poisoned poisonedError
				if !errors.As(err, &poisoned) || !strings.Contains(err.Error(), "differs from a rebuild") {
					t.Fatalf("changed __TEXT: %v", err)
				}
			} else if err != nil {
				t.Fatalf("linker-only change: %v", err)
			}
			after, err := os.ReadFile(filepath.Join(fetched, "product"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("audit changed the fetched file: %v", err)
			}
		})
	}
}

func TestAuditLeavesNonMachOBytesAlone(t *testing.T) {
	t.Parallel()
	for name, content := range map[string][]byte{
		"short": {1, 2, 3},
		"ELF":   append([]byte{0x7f, 'E', 'L', 'F'}, bytes.Repeat([]byte{0xa5}, 128)...),
		"other": bytes.Repeat([]byte{0xa5}, 128),
	} {
		t.Run(name, func(t *testing.T) {
			got, err := auditBytes(content)
			if err != nil || !bytes.Equal(got, content) {
				t.Fatalf("non-Mach-O changed: %x, %v", got, err)
			}
			fetched := t.TempDir()
			write(t, fetched, "product", string(content))
			changed := bytes.Clone(content)
			changed[len(changed)-1] ^= 1
			err = audit(strings.Repeat("b", 64), name, fetched, func(directory string) error {
				return os.WriteFile(filepath.Join(directory, "product"), changed, 0o644)
			})
			var poisoned poisonedError
			if !errors.As(err, &poisoned) {
				t.Fatalf("non-Mach-O mismatch ignored: %v", err)
			}
		})
	}
}

func machOFixture(t *testing.T) []byte {
	t.Helper()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "hello")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-o", binary, "./internal/buildcache/testdata/hello")
	command.Dir = root
	command.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=arm64", "CGO_ENABLED=0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Mach-O fixture: %v\n%s", err, output)
	}
	content, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
