package native_runtime

import (
	"bytes"
	"regexp"
	"strconv"
	"testing"
)

// The IR's type tags reach the runtime's key through adamic.h's bytes, not through the IR: its enum adamic_ir_type is
// one of the files the key hashes, and internal/native's TestIRTypeTags holds it equal to ir.TypeTags, so a tag whose
// number moves in the header moves this machine's key and the portable parts' key alike, with no part of its own
// (#tb722h2 H2). Mutant (ir_tags.mutant.md): runtimeSourceParts passing over the file named adamic.h.
func TestAnIRTagEditedInTheHeaderMovesTheKey(t *testing.T) {
	t.Parallel()
	files, err := ReadRuntime(Sources, "c")
	if err != nil {
		t.Fatal(err)
	}
	tag := regexp.MustCompile(`adamic_ir_[A-Za-z0-9]+[ \t]*=[ \t]*([0-9]+),`)
	edited := append([]RuntimeFile(nil), files...)
	found := false
	for index, file := range edited {
		if file.Name != "adamic.h" {
			continue
		}
		enum := bytes.Index(file.Contents, []byte("enum adamic_ir_type {"))
		if enum < 0 {
			t.Fatal("adamic.h holds no enum adamic_ir_type")
		}
		match := tag.FindSubmatchIndex(file.Contents[enum:])
		if match == nil {
			t.Fatal("adamic.h's enum adamic_ir_type holds no adamic_ir_<Name> = <n> line")
		}
		start, end := enum+match[2], enum+match[3]
		number, err := strconv.Atoi(string(file.Contents[start:end]))
		if err != nil {
			t.Fatal(err)
		}
		contents := append(append([]byte(nil), file.Contents[:start]...), strconv.Itoa(number+1000)...)
		edited[index].Contents = append(contents, file.Contents[end:]...)
		found = true
	}
	if !found {
		t.Fatal("the runtime's files hold no adamic.h")
	}
	flags := []string{"-O2"}
	if runtimeKey(files, flags, "clang", "version", "") == runtimeKey(edited, flags, "clang", "version", "") {
		t.Error("a tag's number edited in adamic.h left this machine's runtime key where it was")
	}
	if PartsKey(PortableKeyParts(files, flags, "clang", "")) == PartsKey(PortableKeyParts(edited, flags, "clang", "")) {
		t.Error("a tag's number edited in adamic.h left the portable parts' key where it was")
	}
}
