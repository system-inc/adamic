package buildcache

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

// An audited product whose build bakes its own build directory into a binary, as a C file compiled where it stands with
// no prefix map does, rebuilds otherwise in the audit's fresh directory, and the audit names the file and the section
// the path landed in, after the mismatch line Loom reads, which stays as it was (#dc483fc). relocatable doesn't read a
// binary's paths, so this is where such a path shows. Mutant: auditHit rebuilding into the stored product's own
// directory, where the rebuild matches: no mismatch, and this test fails.
// Not parallel: newRig.
func TestAnAuditMismatchNamesTheFileAndSectionABuildDirectoryLandedIn(t *testing.T) {
	r := newRig(t)
	witness := Inputs{Name: "witness", Files: []string{"source"}, Toolchain: []string{"clang"}}
	Product(t, witness, bakingItsDirectory(t))
	build := r.built(t)["witness"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	temporary, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", temporary)
	forgetHeldForLife()
	keys := filepath.Join(t.TempDir(), "keys")
	if err := os.WriteFile(keys, []byte(build.NameKey+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_AUDIT_KEYS", keys)
	log := filepath.Join(t.TempDir(), "builds.log")
	t.Setenv("ADAMIC_BUILD_LOG", log)
	if _, err := Get(witness, bakingItsDirectory(t)); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(log)
	lines := string(content)
	t.Logf("census:\n%s", lines)
	if !regexp.MustCompile(`(?m)^audit witness [0-9a-f]{64} mismatch [0-9a-f]{64} [0-9a-f]{64} \S+$`).MatchString(lines) {
		t.Fatalf("a product baking its build directory into a binary: census %q, want a mismatch audit line as Loom reads it", lines)
	}
	if !regexp.MustCompile(`(?m)^audit-difference witness [0-9a-f]{64} witness\.o \(ELF section \.(rodata|debug_str)\S*, at its byte \d+\)$`).MatchString(lines) {
		t.Errorf("census %q, want an audit-difference line naming witness.o and the section holding the path", lines)
	}
}

// bakingItsDirectory builds witness.o in the product's directory from a C file compiled where it stands, with no prefix
// map, so the object's string section holds the build directory. Without clang it writes, by hand, an ELF object
// whose .rodata holds the directory, as the compiled one's does.
func bakingItsDirectory(t *testing.T) func(directory string) error {
	return func(directory string) error {
		object := filepath.Join(directory, "witness.o")
		if clang, err := exec.LookPath("clang"); err == nil {
			source := filepath.Join(directory, "witness.c")
			if err := os.WriteFile(source, []byte("const char *witness = __FILE__;\n"), 0o644); err != nil {
				return err
			}
			command := exec.Command(clang, "-c", "-o", object, source)
			command.Dir = directory
			if output, err := command.CombinedOutput(); err != nil {
				t.Errorf("clang: %v\n%s", err, output)
				return err
			}
			return os.Remove(source)
		}
		return os.WriteFile(object, elfHolding(directory), 0o644)
	}
}

// elfHolding is a minimal ELF64 relocatable object, as debug/elf reads it: a null section, .rodata holding text and a
// NUL, and .shstrtab, with the section headers last.
func elfHolding(text string) []byte {
	names := "\x00.rodata\x00.shstrtab\x00"
	rodata := append([]byte(text), 0)
	var body bytes.Buffer
	body.Write(make([]byte, 64))
	rodataOffset := body.Len()
	body.Write(rodata)
	namesOffset := body.Len()
	body.WriteString(names)
	for body.Len()%8 != 0 {
		body.WriteByte(0)
	}
	sectionsOffset := body.Len()
	type section struct {
		Name, Type    uint32
		Flags, Addr   uint64
		Offset, Size  uint64
		Link, Info    uint32
		Align, Entity uint64
	}
	for _, header := range []section{
		{},
		{Name: 1, Type: uint32(elf.SHT_PROGBITS), Flags: uint64(elf.SHF_ALLOC), Offset: uint64(rodataOffset), Size: uint64(len(rodata)), Align: 1},
		{Name: 9, Type: uint32(elf.SHT_STRTAB), Offset: uint64(namesOffset), Size: uint64(len(names)), Align: 1},
	} {
		binary.Write(&body, binary.LittleEndian, header)
	}
	content := body.Bytes()
	copy(content, []byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), byte(elf.ELFDATA2LSB), byte(elf.EV_CURRENT)})
	binary.LittleEndian.PutUint16(content[16:], uint16(elf.ET_REL))
	binary.LittleEndian.PutUint16(content[18:], uint16(elf.EM_X86_64))
	binary.LittleEndian.PutUint32(content[20:], uint32(elf.EV_CURRENT))
	binary.LittleEndian.PutUint64(content[40:], uint64(sectionsOffset))
	binary.LittleEndian.PutUint16(content[52:], 64)
	binary.LittleEndian.PutUint16(content[58:], 64)
	binary.LittleEndian.PutUint16(content[60:], 3)
	binary.LittleEndian.PutUint16(content[62:], 2)
	return content
}

// firstDifference names a file only one copy holds, a differing byte's offset in a file that is no object, and the
// differing section of a hand-written ELF object, never a header byte the section's length moved; alike copies read
// as "". Mutants: the section lookup skipped (the offset is named instead); the walk stopping at the first file.
func TestTheFirstDifferenceIsAFileAndWhereInIt(t *testing.T) {
	t.Parallel()
	copyWith := func(files map[string][]byte) string {
		directory := t.TempDir()
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(directory, name), content, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return directory
	}
	for _, check := range []struct {
		name            string
		stored, rebuilt map[string][]byte
		want            string
	}{
		{"alike", map[string][]byte{"a": []byte("x")}, map[string][]byte{"a": []byte("x")}, ""},
		{"only stored", map[string][]byte{"a": []byte("x"), "b": nil}, map[string][]byte{"a": []byte("x")}, "b (only in the stored copy)"},
		{"text", map[string][]byte{"a": []byte("x"), "b": []byte("abcd")}, map[string][]byte{"a": []byte("x"), "b": []byte("abXd")}, "b at byte 2"},
		{"object", map[string][]byte{"o": elfHolding("/tmp/one")}, map[string][]byte{"o": elfHolding("/tmp/longer")}, "o (ELF section .rodata, at its byte 5)"},
	} {
		got, err := firstDifference(copyWith(check.stored), copyWith(check.rebuilt))
		if err != nil || got != check.want {
			t.Errorf("%s: %q, %v, want %q", check.name, got, err, check.want)
		}
	}
}
