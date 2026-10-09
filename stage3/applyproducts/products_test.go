package applyproducts

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

func TestProduct_Stage3Adapted10(t *testing.T) { t.Parallel(); Product(t, 0) }
func TestProduct_Stage3Adapted40(t *testing.T) { t.Parallel(); Product(t, 1) }
func TestProduct_Stage3Adapted70(t *testing.T) { t.Parallel(); Product(t, 2) }
func TestProduct_Stage3Adapted99(t *testing.T) { t.Parallel(); Product(t, 3) }

func tracksInputs(t *testing.T, index int) {
	t.Helper()
	repository, err := root()
	if err != nil {
		t.Fatal(err)
	}
	recipe, err := inputs(repository, index)
	if err != nil {
		t.Fatal(err)
	}
	before, err := buildcache.Key(repository, recipe)
	if err != nil {
		t.Fatal(err)
	}
	// Change the real pin in a private input tree. An omitted pin must fail this assertion.
	scratch := t.TempDir()
	for _, name := range recipe.Files {
		source := filepath.Join(repository, name)
		err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return err
			}
			target := filepath.Join(scratch, relative)
			if entry.IsDir() {
				return os.MkdirAll(target, 0755)
			}
			if entry.Type()&os.ModeSymlink != 0 {
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					return err
				}
				return os.Symlink(link, target)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			return os.WriteFile(target, data, info.Mode())
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	unchanged, err := buildcache.Key(scratch, recipe)
	if err != nil {
		t.Fatal(err)
	}
	if unchanged != before {
		t.Fatal("cache location changed product key")
	}
	if err := os.MkdirAll(filepath.Join(scratch, "stage3"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scratch, "stage3/source.json"), []byte("changed pinned TypeScript source"), 0644); err != nil {
		t.Fatal(err)
	}
	after, err := buildcache.Key(scratch, recipe)
	if err != nil {
		t.Fatal(err)
	}
	if after == before {
		t.Fatal("changed source pin reused product key")
	}
}
func TestStage3Adapted10TracksInputs(t *testing.T) { t.Parallel(); tracksInputs(t, 0) }
func TestStage3Adapted40TracksInputs(t *testing.T) { t.Parallel(); tracksInputs(t, 1) }
func TestStage3Adapted70TracksInputs(t *testing.T) { t.Parallel(); tracksInputs(t, 2) }
func TestStage3Adapted99TracksInputs(t *testing.T) { t.Parallel(); tracksInputs(t, 3) }
