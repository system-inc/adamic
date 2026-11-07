// Package graphql holds the test of stage 1's sixth slice: cohere's GraphQL parser, graphql-js's lexer
// and parser as Prettier calls them, ported to Adamic (the .ts files beside this one), compiled natively
// by stage 0 and run on Node, parsing the same texts cohere's own tests parse and thousands more, and
// giving the same tree, comments or syntax error for each, byte for byte as Go cohere does. graphql-js
// throws where a text doesn't parse, and so does the port, from as deep in its recursion as the
// failure is; the one catch is parse's.
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/graphql/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// repository is the repository's root, from this package's directory.
const repository = "../../.."

// portFiles are the port and its driver.
var portFiles = []string{"token.ts", "characterClasses.ts", "blockString.ts", "lexer.ts", "parser.ts", "main.ts"}

// generatedSeed fixes the generated documents, so a failure names documents anyone can make again.
// Setting COHERE_GRAPHQL_SEED asks for others, and COHERE_GRAPHQL_GENERATED for more or fewer of them.
const generatedSeed = 20261005

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port parses every text Go cohere's tests hold and thousands generated as Go cohere does: the same
// tree, field for field and offset for offset, the same comments, or the same syntax error, natively,
// on Node and through the JavaScript backend, byte for byte; and the native port leaks nothing, on the
// texts it parses and on those it throws out of. The sanitized port and each mutant's are portSanitized's,
// built ahead; the lowered program is for the JavaScript backend, and for leaks off Linux.
func TestThePortParsesAsGoCohereDoes(test *testing.T) {
	test.Parallel()
	casesPath, goAnswers := askedCases(test)
	portSource := portDirectory(test, nil)
	program := lowered(test, filepath.Join(portSource, "main.ts"))

	test.Run("natively and on Node, as Go cohere", func(test *testing.T) {
		test.Parallel()
		nodeRun := onNode(test, filepath.Join(portSource, "main.ts"), casesPath)
		sanitized := filepath.Join(portSanitized(test, nil), "port")
		nativeRun := sanitizedRun(test, sanitized, casesPath)
		backendRun := onJavaScriptBackend(test, program, casesPath)
		for _, side := range []struct {
			name string
			run  run
		}{{"Node", nodeRun}, {"native", nativeRun}, {"the JavaScript backend", backendRun}} {
			if side.run.exitCode != 0 || len(side.run.stderr) > 0 {
				test.Fatalf("%s: exit %d, stderr %q", side.name, side.run.exitCode, side.run.stderr)
			}
			if difference := firstDifference(string(side.run.stdout), goAnswers); difference != "" {
				test.Errorf("%s and Go cohere differ: %s", side.name, difference)
			}
		}
		agreed := !test.Failed()
		if leaked := leaks(test, program, sanitized, casesPath); leaked != "" {
			test.Errorf("leaks:\n%s", leaked)
		}

		// Cases that answer the same whatever the parser does would agree without testing it: each kind
		// of node, field and error has to be among them.
		var missing []string
		for _, each := range []string{
			"SchemaDefinition[",
			"ScalarTypeDefinition[",
			"ObjectTypeDefinition[",
			"InterfaceTypeDefinition[",
			"UnionTypeDefinition[",
			"EnumTypeDefinition[",
			"InputObjectTypeDefinition[",
			"DirectiveDefinition[",
			"SchemaExtension[",
			"ScalarTypeExtension[",
			"ObjectTypeExtension[",
			"InterfaceTypeExtension[",
			"UnionTypeExtension[",
			"EnumTypeExtension[",
			"InputObjectTypeExtension[",
			"DirectiveExtension[",
			"FragmentDefinition[",
			"FragmentSpread[",
			"FragmentArgument[",
			"InlineFragment[",
			"VariableDefinition[",
			"ObjectField[",
			"ListValue[",
			"NonNullType[",
			"ListType[",
			"FloatValue[",
			"NullValue[",
			"EnumValue[",
			"BooleanValue[",
			"block=true",
			"repeatable=true",
			"directives=(",
			"values=()",
			"fields=()",
			"\ncomment Comment[",
			`value="\u{1F600}"`,
			`\u{FFFD}`,
			"Invalid Unicode escape sequence",
			"Invalid character escape sequence",
			"Unterminated string.",
			"Unexpected character: U+",
			"Invalid number, expected digit before",
			"Unexpected single quote",
			"Invalid number, unexpected digit after 0",
			"Invalid number, expected digit but got",
			`did you mean "0.`,
			`Unexpected "..", did you mean "..."?`,
			"is reserved and cannot be used for an enum value",
			"in constant value",
			"not supported on shorthand queries",
			"only GraphQL definitions support descriptions",
			"Expected Name, found <EOF>",
		} {
			if !strings.Contains(goAnswers, each) {
				missing = append(missing, each)
			}
		}
		if len(missing) > 0 {
			test.Errorf("the cases never reach %q", missing)
		}
		if agreed {
			cases := strings.Count(goAnswers, "case ")
			refused := strings.Count(goAnswers, "\nerror ")
			test.Logf(
				"%d texts, %d parsed (%d nodes, %d comments), %d refused: every answer the same from Go cohere, the port natively, on Node and through the JavaScript backend",
				cases,
				cases-refused,
				strings.Count(goAnswers, "["),
				strings.Count(goAnswers, "\ncomment "),
				refused,
			)
		}
	})

	// graphql-js 17.0.2 itself, run on every case, which holds Go cohere to the library too.
	test.Run("as graphql-js", func(test *testing.T) {
		test.Parallel()
		library := os.Getenv("ADAMIC_GRAPHQL_LIBRARY")
		if library == "" {
			test.Skip(
				"set ADAMIC_GRAPHQL_LIBRARY to a directory where `npm install graphql@17.0.2` ran, to compare with the library itself",
			)
		}
		script, absolutePathError := filepath.Abs(filepath.Join("testdata", "library.mjs"))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		libraryRun := execute(test, nil, "node", script, library, casesPath)
		if libraryRun.exitCode != 0 || len(libraryRun.stderr) > 0 {
			test.Fatalf("the library: exit %d, stderr %q", libraryRun.exitCode, libraryRun.stderr)
		}
		if difference := firstDifference(string(libraryRun.stdout), goAnswers); difference != "" {
			test.Errorf(
				"graphql-js and Go cohere differ: %s",
				strings.Replace(difference, "line", "graphql-js's line", 1),
			)
		}
	})

	// The comparison has to be able to fail. Each mutant changes the port where only its answers can
	// show it, and the port must then disagree with Go cohere: natively, since that is the program this
	// test holds, built ahead as portSanitized's, and on Node, which shows the fault is the port's and
	// not stage 0's.
	for _, mutant := range mutants {
		test.Run("catches "+mutant.name, func(test *testing.T) {
			test.Parallel()
			mutated := portDirectory(test, &mutant)
			binary := filepath.Join(portSanitized(test, &mutant), "port")
			for _, side := range []struct {
				name string
				run  run
			}{
				{"natively", sanitizedRun(test, binary, casesPath)},
				{"on Node", onNode(test, filepath.Join(mutated, "main.ts"), casesPath)},
			} {
				if side.run.exitCode != 0 {
					test.Errorf(
						"%s the mutant exits %d (stderr %q); it must be caught by its answers, not by failing",
						side.name,
						side.run.exitCode,
						side.run.stderr,
					)
					continue
				}
				difference := firstDifference(string(side.run.stdout), goAnswers)
				if difference == "" {
					test.Errorf("%s the mutant agrees with Go cohere: the comparison cannot see it", side.name)
					continue
				}
				test.Logf("%s, caught: %s", side.name, difference)
			}
		})
	}
}

// mutant is one change to one port file.
type mutant struct {
	name, file, from, to string
}

// mutants are the mutants the comparison must catch; TestProduct_GraphQLPortMutantsSanitized builds each ahead.
var mutants = []mutant{
	// An edge of a class: the byte order mark is ignored between tokens, as graphql-js ignores it.
	{
		name: "a byte order mark not ignored",
		file: "lexer.ts",
		from: "            case 0xfeff: // <BOM>\n",
		to:   "",
	},
	// An edge of a class: a no-break space is no separator, but "Unexpected character: U+00A0.".
	{
		name: "a no-break space ignored",
		file: "lexer.ts",
		from: "            case 0xfeff: // <BOM>\n",
		to:   "            case 0xfeff: // <BOM>\n\t\t\tcase 0x00a0:\n",
	},
	// An edge of a class: `[`, just past `Z`, taken for a letter, so `a[` is one name.
	{
		name: "a [ in a name",
		file: "characterClasses.ts",
		from: "(code >= 0x0041 && code <= 0x005a) // A-Z",
		to:   "(code >= 0x0041 && code <= 0x005b) // A-Z",
	},
	// An edge of a class: `G`, just past `F`, taken for a hex digit.
	{
		name: "a G in a hex escape",
		file: "lexer.ts",
		from: ": code >= 0x0041 && code <= 0x0046",
		to:   ": code >= 0x0041 && code <= 0x0047",
	},
	// The edge of graphql-js's 32-bit arithmetic: a point that reaches the sign bit stops the escape a
	// character early, and the message quotes one character less.
	{
		name: "the sign bit not reached",
		file: "lexer.ts",
		from: "point * 16 + digit > 0x7fffffff ? -1",
		to:   "point * 16 + digit > 0xffffffff ? -1",
	},
	// An edge of a class: U+DFFF is the last trailing surrogate, so the escapes of U+DBFF and U+DFFF spell U+10FFFF.
	{
		name: "U+DFFF no trailing surrogate",
		file: "lexer.ts",
		from: "return code >= 0xdc00 && code <= 0xdfff;",
		to:   "return code >= 0xdc00 && code < 0xdfff;",
	},
	// A lone surrogate, which a message's slice through a surrogate pair makes, written as itself rather
	// than as the U+FFFD Node writes for it.
	{
		name: "a lone surrogate written as itself",
		file: "lexer.ts",
		from: "    if(code >= 0xd800 && code <= 0xdfff) {\n        return '\\\\u{FFFD}';\n    }\n",
		to:   "",
	},
	// A carriage return and line feed counted as two lines in an error's location.
	{
		name: "a CRLF counted as two lines",
		file: "lexer.ts",
		from: "const length = code === 0x000d && body.charCodeAt(index + 1) === 0x000a ? 2 : 1;",
		to:   "const length = 1;",
	},
	// A block string's first line counted in the common indent.
	{
		name: "the first line's indent common",
		file: "blockString.ts",
		from: "if(index !== 0 && indent < commonIndent) {",
		to:   "if(indent < commonIndent) {",
	},
	// The backslash of an escaped triple quote kept in the value.
	{
		name: "an escaped triple quote keeping its backslash",
		file: "lexer.ts",
		from: "chunkStart = position + 1; // skip only slash",
		to:   "chunkStart = position; // skip only slash",
	},
	// A thrown error's position: a digit after a leading 0 reported at the number's start.
	{
		name: "a leading zero's error at the number's start",
		file: "lexer.ts",
		from: "throw new Error(\n                syntaxError(\n                    body,\n                    position,\n                    `Invalid number, unexpected digit after 0:",
		to:   "throw new Error(\n                syntaxError(\n                    body,\n                    start,\n                    `Invalid number, unexpected digit after 0:",
	},
	// A throw not taken: an escape graphql-js refuses read as nothing, and the parse goes on.
	{
		name: "an invalid escape read as nothing",
		file: "lexer.ts",
		from: "    throw new Error(\n        syntaxError(\n            body,\n            position,\n            `Invalid character escape sequence: \"${escaped(body.slice(position, position + 2))}\".`,\n        ),\n    );",
		to:   "    return { value: '', size: 2 };",
	},
	// unexpected told which token to name, naming the current one instead (`extend thing X` names
	// `thing`, the lookahead).
	{
		name: "the unexpected token not the one named",
		file: "parser.ts",
		from: "const token = atToken ?? this.lexer.token;",
		to:   "const token = this.lexer.token ?? atToken;",
	},
	// No directives written as an empty list rather than undefined.
	{
		name: "no directives an empty list",
		file: "parser.ts",
		from: "        if(directives.length > 0) {\n            return { kind: 'List', nodes: directives };\n        }\n        return undefinedValue;",
		to:   "            return { kind: 'List', nodes: directives };",
	},
	// A fragment spread without arguments written with arguments undefined, where graphql-js never
	// writes the field.
	{
		name: "a spread's absent arguments written",
		file: "parser.ts",
		from: "                            this.node(fragmentStart, 'FragmentSpread', [\n                                child('name', name),\n                                field('directives', this.parseDirectives(false)),\n                            ]),",
		to:   "                            this.node(fragmentStart, 'FragmentSpread', [\n                                child('name', name),\n                                field('arguments', undefinedValue),\n                                field('directives', this.parseDirectives(false)),\n                            ]),",
	},
	// One of the 21 directive locations forgotten.
	{
		name: "FRAGMENT_VARIABLE_DEFINITION no location",
		file: "parser.ts",
		from: "        case 'FRAGMENT_VARIABLE_DEFINITION':\n",
		to:   "",
	},
}

// askedCases has cohere's side write every case, and Go cohere's answer to each, returning the cases
// file's path and the answers.
func askedCases(test *testing.T) (string, string) {
	test.Helper()
	seed := int64(generatedSeed)
	var parseError error
	if value := os.Getenv("COHERE_GRAPHQL_SEED"); value != "" {
		if seed, parseError = strconv.ParseInt(value, 10, 64); parseError != nil {
			test.Fatal(parseError)
		}
	}
	generated := 3000
	if value := os.Getenv("COHERE_GRAPHQL_GENERATED"); value != "" {
		if generated, parseError = strconv.Atoi(value); parseError != nil {
			test.Fatal(parseError)
		}
	}
	directory := test.TempDir()
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	cohereSide(test, map[string]any{"seed": seed, "generated": generated, "cases": casesPath, "answers": answersPath})
	answers, readError := os.ReadFile(answersPath)
	if readError != nil {
		test.Fatal(readError)
	}
	if keep := os.Getenv("ADAMIC_GRAPHQL_KEEP"); keep != "" {
		cases, readError := os.ReadFile(casesPath)
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(keep, cases, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	test.Logf("seed %d, %d generated documents", seed, generated)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's graphql package, by overlay, with a
// request.
func cohereSide(test *testing.T, request map[string]any) {
	test.Helper()
	directory := test.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, encodeError := json.Marshal(request)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(requestPath, encoded, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	// The child guard owns hang detection.
	binary, packageDirectory := cohereSideOracle(test)
	command := bounded(test, binary, "-test.timeout=0", "-test.run=^TestAdamicPortCases$")
	command.Dir = packageDirectory
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, combinedOutputError := combinedOutput(command); combinedOutputError != nil {
		test.Fatalf("cohere's side: %v\n%s", combinedOutputError, output)
	}
}

// portDirectory copies the port into a directory of its own, with a mutant applied when there is one.
func portDirectory(test *testing.T, applied *mutant) string {
	test.Helper()
	directory := test.TempDir()
	for _, name := range portFiles {
		contents, readError := os.ReadFile(name)
		if readError != nil {
			test.Fatal(readError)
		}
		source := string(contents)
		if applied != nil && applied.file == name {
			if strings.Count(source, applied.from) != 1 {
				test.Fatalf("the mutant %q must change exactly one place in %s", applied.name, name)
			}
			source = strings.Replace(source, applied.from, applied.to, 1)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return directory
}

// firstDifference says where two outputs first differ, line by line, or "" when they don't.
func firstDifference(got string, want string) string {
	if got == want {
		return ""
	}
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want, "\n")
	for index := 0; index < len(gotLines) || index < len(wantLines); index++ {
		var gotLine, wantLine string
		if index < len(gotLines) {
			gotLine = gotLines[index]
		}
		if index < len(wantLines) {
			wantLine = wantLines[index]
		}
		if gotLine != wantLine {
			return fmt.Sprintf("line %d: %q, Go cohere %q", index+1, gotLine, wantLine)
		}
	}
	return "the same lines, not the same bytes"
}

// lowered checks and lowers a program, failing the test with stage 0's refusal if it can't.
func lowered(test *testing.T, path string) *ir.Program {
	test.Helper()
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatalf("Load: %v", loadError)
	}
	result, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatalf("Lower: %v", lowerError)
	}
	return result
}

// bounded prepares a child; execute and combinedOutput run it with progress-based guards.
func bounded(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	// The child guard owns hang detection, including for Go test children.
	if name == "go" && len(arguments) > 0 && arguments[0] == "test" {
		arguments = append([]string{"test", "-timeout=0"}, arguments[1:]...)
	}
	return exec.Command(name, arguments...)
}

func combinedOutput(command *exec.Cmd) ([]byte, error) {
	return childguard.CombinedOutput(command, childguard.Options{})
}

func execute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	command := bounded(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runError := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if runError != nil && !errors.As(runError, &exitError) {
		test.Fatalf("running %s: %v", name, runError)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return execute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(test *testing.T, program *ir.Program, arguments ...string) run {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.mjs")
	if writeError := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	return onNode(test, path, arguments...)
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(test *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	test.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(test.TempDir(), "port")
		if buildError := native.Build(native.C(program), binary, native.Options{}); buildError != nil {
			test.Fatal(buildError)
		}
		report := execute(test, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// cohereSideOracle is the test binary go test would build with testdata/cohere_side_test.go laid over cohere's code, a
// product (buildcache.GoTest) built ahead, and the directory go test would run it in.
func cohereSideOracle(test testing.TB) (string, string) { return oracles.Parser(test) }

func TestProduct_GraphQLFindsGoOracle(test *testing.T) {
	test.Parallel()
	cohereSideOracle(test)
}
