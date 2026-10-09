package rule

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// What one program's design system read survives another program asking for its own. "Not loaded" has to
// mean no rule asked on this program, because the findings cache records a program whose design system was
// never loaded as independent of every stylesheet. When this was one slot, a second program took it and the
// first then read as never loaded (#35nqkwc, @system_cohere_lint's review).
func TestDesignSystemReadsAreKeptPerProgram(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	theme := filepath.Join(directory, "theme.css")
	if err := os.WriteFile(theme, []byte("@theme {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	first, second, neither := &compiler.Program{}, &compiler.Program{}, &compiler.Program{}

	firstRecorder := NewRecordingFS(osvfs.FS())
	recordDesignSystemFS(first, firstRecorder)
	firstRecorder.ReadFile(tspath.RootedFilePath(theme))
	secondRecorder := NewRecordingFS(osvfs.FS())
	recordDesignSystemFS(second, secondRecorder)
	secondRecorder.FileExists(tspath.RootedFilePath(filepath.Join(directory, "missing.css")))

	reads, loaded := DesignSystemReads(first)
	if !loaded || len(reads) != 1 || reads[0].Path != theme || !reads[0].Present {
		t.Errorf("the first program's reads after a second program asked: %v, loaded %v; want theme.css present", reads, loaded)
	}
	reads, loaded = DesignSystemReads(second)
	if !loaded || len(reads) != 1 || reads[0].Present {
		t.Errorf("the second program's reads: %v, loaded %v; want missing.css absent", reads, loaded)
	}
	if reads, loaded := DesignSystemReads(neither); loaded || len(reads) != 0 {
		t.Errorf("a program that never asked reads as loaded: %v", reads)
	}
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
}

// An entry does not keep its program alive, and an entry whose program is gone is swept on the next
// registration, so a test process that builds thousands of programs does not hold them all.
func TestDesignSystemReadsDoNotHoldTheirProgram(t *testing.T) {
	t.Parallel()
	gone := weak.Make(func() *compiler.Program {
		program := &compiler.Program{}
		recordDesignSystemFS(program, NewRecordingFS(osvfs.FS()))
		return program
	}())
	for attempt := 0; attempt < 5 && gone.Value() != nil; attempt++ {
		runtime.GC()
	}
	if gone.Value() != nil {
		t.Fatal("a program held only by its design system entry was not collected")
	}
	kept := &compiler.Program{}
	recordDesignSystemFS(kept, NewRecordingFS(osvfs.FS()))
	designSystemRecordings.Lock()
	_, stillThere := designSystemRecordings.byProgram[gone]
	designSystemRecordings.Unlock()
	if stillThere {
		t.Error("the entry of a collected program was not swept by the next registration")
	}
	runtime.KeepAlive(kept)
}
