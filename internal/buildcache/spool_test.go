package buildcache

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func runPublisher(t *testing.T) {
	t.Helper()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", "./internal/buildcache/cmd/buildcache-publish")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("publish: %v\n%s", err, output)
	}
}

func TestConcurrentPublishersDrainOnceAndRetryIsEmpty(t *testing.T) {
	store, _ := shared(t)
	token := filepath.Join(t.TempDir(), "publish-token")
	write(t, filepath.Dir(token), filepath.Base(token), "gate-box-token")
	t.Setenv("ADAMIC_BUILD_STORE_TOKEN", token)
	Product(t, thisPackage, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	})
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			if err := PublishSpool(); err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
	if len(store.writes) != 3 {
		t.Fatalf("writes: %v", store.writes)
	}
	if err := PublishSpool(); err != nil {
		t.Fatal(err)
	}
	if len(store.writes) != 3 {
		t.Fatalf("retry wrote %v", store.writes)
	}
}

func TestAnUploadFailureKeepsItsSpoolEntry(t *testing.T) {
	store, _ := shared(t)
	token := filepath.Join(t.TempDir(), "publish-token")
	write(t, filepath.Dir(token), filepath.Base(token), "wrong-token")
	t.Setenv("ADAMIC_BUILD_STORE_TOKEN", token)
	Product(t, thisPackage, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	})
	if err := PublishSpool(); err == nil {
		t.Fatal("an upload failure was ignored")
	}
	entries, err := spoolEntries()
	if err != nil || len(entries) != 1 {
		t.Fatalf("failed upload's spool: %v, %v", entries, err)
	}
	write(t, filepath.Dir(token), filepath.Base(token), "gate-box-token")
	if err := PublishSpool(); err != nil {
		t.Fatal(err)
	}
	entries, err = spoolEntries()
	if err != nil || len(entries) != 0 {
		t.Fatalf("retry's spool: %v, %v", entries, err)
	}
	if _, ok := store.objects["/refs/build-candidate/"+thisKey(t)]; !ok {
		t.Fatal("retry didn't publish its ref")
	}
}
