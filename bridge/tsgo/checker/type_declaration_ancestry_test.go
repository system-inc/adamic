package checker

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestTypeDeclarationAncestry(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "type Outcome<T> = (({outcome:'Ok'; value:T}) | ({outcome:'Error'; message:string}));\ndeclare function sync():Outcome<number>;\ndeclare function async():Promise<Outcome<number>>;\nsync();\nasync();\n"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	ask := func(name, mode string) []string {
		t.Helper()
		end := strings.LastIndex(source, "\n"+name+"()") + 1
		start := end - 1
		wire, err := program.Inspect(file, uint64(start), uint64(end+len(name)+2), "CallExpression", "type-declaration-ancestry\n"+mode)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	inspect := func(answer []string) map[string]int {
		t.Helper()
		count, err := strconv.Atoi(answer[2])
		if err != nil {
			t.Fatal(err)
		}
		if count == 0 {
			t.Fatal("missing symbol declarations")
		}
		position := 3 + count
		length, err := strconv.Atoi(answer[position])
		if err != nil {
			t.Fatal(err)
		}
		position++
		if len(answer) != position+length*7 {
			t.Fatal("wrong ancestry wire length")
		}
		ids := map[string]bool{}
		kinds := map[string]int{}
		for index := 0; index < length; index++ {
			record := answer[position+index*7 : position+(index+1)*7]
			if ids[record[0]] || record[0] == "0" {
				t.Fatal("invalid node identity")
			}
			ids[record[0]] = true
			kinds[record[2]]++
		}
		for index := 0; index < length; index++ {
			record := answer[position+index*7 : position+(index+1)*7]
			for _, identity := range []string{record[1], record[5]} {
				if identity != "0" && !ids[identity] {
					t.Fatal("missing link")
				}
			}
			if record[2] == "TypeAliasDeclaration" && (record[3] != "Outcome" || record[4] != file || record[5] == "0") {
				t.Fatalf("lost alias identity: %q", record)
			}
			if record[2] == "UnionType" && record[6] != "2" {
				t.Fatal("lost union arm count")
			}
		}
		return kinds
	}
	raw := ask("sync", "raw")
	awaited := ask("async", "awaited")
	for _, answer := range [][]string{raw, awaited} {
		kinds := inspect(answer)
		if kinds["TypeLiteral"] != 2 || kinds["UnionType"] != 1 || kinds["TypeAliasDeclaration"] != 1 || kinds["ParenthesizedType"] != 3 || kinds["SourceFile"] != 1 {
			t.Fatalf("lost ancestry: %v", kinds)
		}
	}
	floating := inspect(ask("async", "raw"))
	if floating["TypeAliasDeclaration"] != 0 || floating["InterfaceDeclaration"] == 0 {
		t.Fatal("floating promise treated as its outcome")
	}
	end := strings.LastIndex(source, "\nsync()") + 1
	for _, question := range []string{"type-declaration-ancestry", "type-declaration-ancestry\n", "type-declaration-ancestry\nraw\nextra", "type-declaration-ancestry\ninvalid", "unknown"} {
		if _, err := program.Inspect(file, uint64(end-1), uint64(end+6), "CallExpression", question); err == nil {
			t.Fatal("accepted malformed ancestry question", question)
		}
	}
}
