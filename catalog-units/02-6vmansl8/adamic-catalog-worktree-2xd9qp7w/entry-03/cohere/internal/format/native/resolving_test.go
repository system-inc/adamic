package native

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatoptions"
)

// TestResolvingFormatsEachFileWithItsOwnDirectorysOptions: two repositories with different configs,
// one formatter. Formatting api-phi-health with ahra's options is the bug prettier.Resolving exists to
// prevent, and the native engine must not reintroduce it at the switch.
func TestResolvingFormatsEachFileWithItsOwnDirectorysOptions(t *testing.T) {
	t.Parallel()
	Register(".resolvingprobe", func(fileName string, text string, options formatoptions.Options) (string, error) {
		return strconv.Itoa(options.PrintWidth), nil
	})
	defer func() { mutex.Lock(); delete(printers, ".resolvingprobe"); mutex.Unlock() }()

	root := t.TempDir()
	for name, printWidth := range map[string]string{"narrow": "40", "wide": "160"} {
		directory := filepath.Join(root, name, "source")
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		// Each repository extends its own Nexus tier, the one file a format block may live in.
		nexusTier := `{"format": {"printWidth": ` + printWidth + `}}`
		if err := os.WriteFile(filepath.Join(root, name, "NexusCohereSettings.json"), []byte(nexusTier), 0o644); err != nil {
			t.Fatal(err)
		}
		settings := `{"extends": "./NexusCohereSettings.json"}`
		if err := os.WriteFile(filepath.Join(root, name, "CohereSettings.json"), []byte(settings), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Started in narrow, so a formatter that kept its starting options would print 40 for both.
	resolving, err := NewResolving(filepath.Join(root, "narrow"))
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"narrow": "40", "wide": "160"} {
		got, err := resolving.Format(filepath.Join(root, name, "source", "file.resolvingprobe"), "")
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s formatted with printWidth %s, want %s from its own Nexus tier", name, got, want)
		}
	}
}

// TestResolvingOffersOnlyWhatAPrinterHandles: the walk counts a type with no printer as declined,
// by extension, rather than offering it and failing on it or dropping it without a word.
func TestResolvingOffersOnlyWhatAPrinterHandles(t *testing.T) {
	t.Parallel()
	Register(".offeredprobe", func(fileName string, text string, options formatoptions.Options) (string, error) {
		return text, nil
	})
	defer func() { mutex.Lock(); delete(printers, ".offeredprobe"); mutex.Unlock() }()

	root := t.TempDir()
	for _, name := range []string{"kept.offeredprobe", "declined.unregistered"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	resolving, err := NewResolving(root)
	if err != nil {
		t.Fatal(err)
	}
	enumeration, err := resolving.Enumerate(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(enumeration.Files) != 1 || filepath.Base(enumeration.Files[0]) != "kept.offeredprobe" {
		t.Fatalf("offered %v, want only kept.offeredprobe", enumeration.Files)
	}
	if enumeration.DeclinedExtensions[".unregistered"] != 1 {
		t.Fatalf("declined %v, want .unregistered counted once", enumeration.DeclinedExtensions)
	}
}

// TestResolvingRefusesAConfigItCannotApplyAtConstruction: a bad config stops the run before any
// phase works, not on the first file after fixes have begun.
func TestResolvingRefusesAConfigItCannotApplyAtConstruction(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"prettier": "@company/prettier-config"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewResolving(root); err == nil {
		t.Fatal("NewResolving accepted Prettier config left in package.json")
	}
}
