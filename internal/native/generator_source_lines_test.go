package native

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// debugGeneratorFixture's generator has a statement on every line from 2 to 9: a yield at 3 and
// at 6, and a throw at 8.
const debugGeneratorFixture = `function* steps(limit: number) {
    const first = limit + 1;
    yield first;
    for(let index = 0; index < limit; index++) {
        const next = index * 2;
        yield next;
    }
    if(limit > 3) throw new Error('too many');
    console.log('finished');
}
for(const step of steps(2)) {
    console.log(String(step));
}
`

// Under --debug a generator's resume carries a #line for every line of its body, as an async
// function's does (TestDebugAsyncTerminals), since both are emitted by suspensionBlocks with the
// body's source lines, and each yield's terminal (the hand-over, the kill set, the state and the
// return) sits under its yield's own line, so a debugger stopped there reads the .a line of the
// yield (Diagnostics' should 7 on #qk8rztp). This reads the emitted C. The line table a debugger
// reads is clang's from these directives; on the Mac a linked binary keeps no DWARF of its own (it
// stays in the objects), and the toolchain has no llvm-dwarfdump, so the line table itself is the
// box's to read. Mutant: generatorFunction handing suspensionBlocks no source lines; no #line in
// the resume, and this goes red.
func TestDebugGeneratorSourceLines(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "fixture.a")
	if err := os.WriteFile(path, []byte(debugGeneratorFixture), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := debugProgram(t, path, true)
	if err != nil {
		t.Fatal(err)
	}
	code := CWithOptions(program, Options{Debug: true})
	start := strings.Index(code, "static void adamic_generator_resume_")
	if start < 0 {
		t.Fatal("no resume function")
	}
	resume := code[start:]
	resume = resume[:strings.Index(resume, "\n}\n")]
	directive := regexp.MustCompile(`^\s*#line (\d+) "([^"]*)"`)
	lines := map[int]bool{}
	current := 0
	var yields []int
	for _, text := range strings.Split(resume, "\n") {
		if match := directive.FindStringSubmatch(text); match != nil {
			current, _ = strconv.Atoi(match[1])
			if filepath.Base(match[2]) == "fixture.a" {
				lines[current] = true
			}
			continue
		}
		if strings.Contains(text, "frame->status = adamic_generator_status_suspended;") {
			yields = append(yields, current)
		}
	}
	for line := 2; line <= 9; line++ {
		if line == 7 {
			continue // The for loop's closing brace holds no statement.
		}
		if !lines[line] {
			t.Errorf("the resume has no #line for fixture.a:%d (it has %v)", line, lines)
		}
	}
	if len(yields) != 2 || yields[0] != 3 || yields[1] != 6 {
		t.Errorf("the yield terminals sit under lines %v, want [3 6]", yields)
	}
}
