package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// A split is a post-processing prototype of C's output. It deliberately understands only the
// emitter's declarations, not arbitrary C. Unknown syntax fails before invoking clang.
type compilationUnit struct{ name, source string }
type cToken struct {
	text       string
	start, end int
}

// cTokens keeps strings, characters, comments and directives out of the structural scan.
func cTokens(source string) ([]cToken, error) {
	var tokens []cToken
	for i := 0; i < len(source); {
		start := i
		c := source[i]
		if strings.ContainsRune(" \t\r\n", rune(c)) {
			i++
			continue
		}
		if c == '#' {
			for i < len(source) && source[i] != '\n' {
				i++
			}
			tokens = append(tokens, cToken{source[start:i], start, i})
			continue
		}
		if c == '/' && i+1 < len(source) && source[i+1] == '/' {
			for i < len(source) && source[i] != '\n' {
				i++
			}
			continue
		}
		if c == '/' && i+1 < len(source) && source[i+1] == '*' {
			end := strings.Index(source[i+2:], "*/")
			if end < 0 {
				return nil, fmt.Errorf("native: split: unterminated comment")
			}
			i += end + 4
			continue
		}
		if c == '"' || c == '\'' {
			i++
			for i < len(source) && source[i] != c {
				if source[i] == '\\' {
					i++
				}
				i++
			}
			if i >= len(source) {
				return nil, fmt.Errorf("native: split: unterminated literal")
			}
			i++
		} else if identifierByte(c) {
			for i < len(source) && identifierByte(source[i]) {
				i++
			}
		} else {
			i++
		}
		tokens = append(tokens, cToken{source[start:i], start, i})
	}
	return tokens, nil
}
func identifierByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

type cDeclaration struct {
	tokens            []cToken
	body, initializer int
	function          bool
	name              string
}

func splitDeclarations(source string) ([]cDeclaration, error) {
	tokens, err := cTokens(source)
	if err != nil {
		return nil, err
	}
	var declarations []cDeclaration
	for begin := 0; begin < len(tokens); {
		if strings.HasPrefix(tokens[begin].text, "#") {
			declarations = append(declarations, cDeclaration{tokens: tokens[begin : begin+1], body: -1, initializer: -1})
			begin++
			continue
		}
		depth, paren, bracket := 0, 0, 0
		declaration := cDeclaration{body: -1, initializer: -1}
		end := begin
		for ; end < len(tokens); end++ {
			token := tokens[end].text
			if depth == 0 && paren == 0 && bracket == 0 {
				if token == "=" {
					declaration.initializer = end - begin
				}
				if token == "{" && declaration.initializer < 0 && end > begin && tokens[end-1].text == ")" {
					declaration.body = end - begin
					declaration.function = true
				}
				if token == ";" {
					end++
					break
				}
			}
			switch token {
			case "(":
				paren++
			case ")":
				paren--
			case "[":
				bracket++
			case "]":
				bracket--
			case "{":
				depth++
			case "}":
				depth--
			}
			if depth < 0 || paren < 0 || bracket < 0 {
				return nil, fmt.Errorf("native: split: unbalanced declaration")
			}
			if token == "}" && depth == 0 && declaration.function {
				end++
				break
			}
		}
		if end == begin || end > len(tokens) || depth != 0 || paren != 0 || bracket != 0 {
			return nil, fmt.Errorf("native: split: incomplete declaration")
		}
		declaration.tokens = tokens[begin:end]
		limit := len(declaration.tokens)
		if declaration.body >= 0 {
			limit = declaration.body
		}
		if declaration.initializer >= 0 {
			limit = declaration.initializer
		}
		// Generated declarators are a single identifier, followed by an array or argument list.
		for j := 1; j < limit; j++ {
			if declaration.tokens[j].text == "(" {
				declaration.name = declaration.tokens[j-1].text
				break
			}
			if declaration.tokens[j].text == "[" {
				declaration.name = declaration.tokens[j-1].text
				break
			}
		}
		if declaration.name == "" && limit >= 2 {
			declaration.name = declaration.tokens[limit-1].text
		}
		if declaration.name == ";" && limit >= 3 {
			declaration.name = declaration.tokens[limit-2].text
		}
		if declaration.name != "main" && (!cName.MatchString(declaration.name) || declaration.tokens[0].text != "static") {
			return nil, fmt.Errorf("native: split: unsupported declaration near %q", declaration.tokens[0].text)
		}
		declarations = append(declarations, declaration)
		begin = end
	}
	return declarations, nil
}

// splitC gives every generated static symbol one external definition. Token replacement never
// touches literal bytes. The fixed namespace cannot collide with runtime API names. Existing
// program-wide indexes distinguish symbols; no whole-program content hash invalidates all units.
func splitC(source string) (string, []compilationUnit, error) {
	declarations, err := splitDeclarations(source)
	if err != nil {
		return "", nil, err
	}
	names := map[string]string{}
	for _, d := range declarations {
		if d.name != "" && d.name != "main" {
			names[d.name] = "adamic_unit_" + d.name
		}
	}
	rewrite := func(tokens []cToken) string {
		if len(tokens) == 0 {
			return ""
		}
		var out strings.Builder
		position := tokens[0].start
		for _, t := range tokens {
			out.WriteString(source[position:t.start])
			name, found := names[t.text]
			if found {
				out.WriteString(name)
			} else {
				out.WriteString(t.text)
			}
			position = t.end
		}
		return out.String()
	}
	var header, state strings.Builder
	header.WriteString("#ifndef ADAMIC_UNITS_H\n#define ADAMIC_UNITS_H\n")
	var units []compilationUnit
	defined := map[string]bool{}
	for _, d := range declarations {
		if d.name == "" {
			header.WriteString(rewrite(d.tokens) + "\n")
			continue
		}
		tokens := d.tokens
		if tokens[0].text == "static" {
			tokens = tokens[1:]
		}
		if d.function {
			// body indexes still refer to the original tokens, before static was removed.
			signature := rewrite(d.tokens[1:d.body])
			if d.name == "main" {
				signature = rewrite(d.tokens[:d.body])
			}
			header.WriteString(signature + ";\n")
			units = append(units, compilationUnit{d.name + ".c", rewrite(tokens) + "\n"})
			continue
		}
		limit := len(d.tokens) - 1
		if d.initializer >= 0 {
			limit = d.initializer
		}
		header.WriteString("extern " + rewrite(d.tokens[1:limit]) + ";\n")
		// Class forward declarations precede their initialized definition. A slot cache without an
		// initializer is a real zero-initialized definition, so preserve it in the state unit.
		if d.initializer < 0 && strings.HasPrefix(d.name, "adamic_class_") {
			continue
		}
		if defined[d.name] {
			return "", nil, fmt.Errorf("native: split: duplicate definition %s", d.name)
		}
		defined[d.name] = true
		state.WriteString(rewrite(tokens) + "\n")
	}
	header.WriteString("#endif\n")
	// Sixteen consecutive functions amortize clang startup and repeated header parsing. A body
	// edit changes one group; main stays separate from the shared state and function groups.
	var groups []compilationUnit
	functionsInGroup := 0
	for _, unit := range units {
		if unit.name == "main.c" {
			continue
		}
		index := len(groups) - 1
		if index < 0 || functionsInGroup == 16 {
			groups = append(groups, compilationUnit{name: fmt.Sprintf("functions_%04d.c", len(groups))})
			index++
			functionsInGroup = 0
		}
		groups[index].source += unit.source
		functionsInGroup++
	}
	for _, unit := range units {
		if unit.name == "main.c" {
			groups = append(groups, unit)
		}
	}
	units = append([]compilationUnit{{"state.c", state.String()}}, groups...)
	for i := range units {
		units[i].source = "#include \"units.h\"\n" + units[i].source
	}
	return header.String(), units, nil
}

var unitBuilds sync.Map

// Preprocessed source/header bytes, ordered flags, compiler identity/version and platform enter
// the key. Preprocessing on every lookup also observes ambient and system include dependencies.
func unitKey(files []runtimeFile, flags []string, compiler, version string) string {
	return runtimeKey(files, append([]string{"adamic-units-v2"}, flags...), compiler, version)
}

func compileUnit(unit compilationUnit, files []runtimeFile, flags []string, compiler, version, cache, directory string, uncached bool) (string, error) {
	// Preprocess the snapshotted source on every lookup: the result includes transitive system
	// headers, ambient include paths, target macros and time-dependent predefined macros. Hash
	// those exact bytes and compile that same snapshot, rather than reopening dependencies.
	// Keep clang line markers: their system-header flag suppresses extension diagnostics inside
	// platform headers under -pedantic -Werror. Flattening with -P loses that provenance.
	temporary, err := os.MkdirTemp(directory, ".unit-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(temporary)
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(temporary, file.name), file.contents, 0o644); err != nil {
			return "", err
		}
	}
	preprocess := exec.Command(compiler, append(append([]string{}, flags...), "-E", unit.name)...)
	preprocess.Dir = temporary
	var diagnostic strings.Builder
	preprocess.Stderr = &diagnostic
	preprocessed, err := preprocess.Output()
	if err != nil {
		return "", fmt.Errorf("native: preprocessing unit %s: %w\n%s", unit.name, err, &diagnostic)
	}
	key := unitKey([]runtimeFile{{unit.name, preprocessed}}, flags, compiler, version)
	entry := filepath.Join(cache, key)
	object := filepath.Join(entry, "unit.o")
	value, _ := unitBuilds.LoadOrStore(entry, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	if !uncached {
		if info, err := os.Stat(object); err == nil && info.Mode().IsRegular() {
			return object, nil
		}
	}
	if err := os.WriteFile(filepath.Join(temporary, unit.name), preprocessed, 0o644); err != nil {
		return "", err
	}
	// Macro options already took effect during preprocessing. clang warns when they are passed
	// to a preprocessed input. They remain in the key, in their original order.
	compileFlags := []string{}
	for _, flag := range flags {
		if !strings.HasPrefix(flag, "-D") && !strings.HasPrefix(flag, "-U") {
			compileFlags = append(compileFlags, flag)
		}
	}
	// Clang emits GNU line-marker syntax even for C11. Accept only that generated syntax;
	// keep all other pedantic diagnostics, including extensions outside system headers.
	arguments := append(append([]string{}, compileFlags...), "-Wno-gnu-line-marker", "-fdebug-prefix-map="+temporary+"=/adamic-units", "-x", "cpp-output", "-c", unit.name, "-o", "unit.o")
	command := exec.Command(compiler, arguments...)
	command.Dir = temporary
	if output, err := command.CombinedOutput(); err != nil {
		return "", fmt.Errorf("native: compiling unit %s: %w\n%s", unit.name, err, output)
	}
	if uncached {
		object = filepath.Join(directory, key+".o")
		return object, os.Rename(filepath.Join(temporary, "unit.o"), object)
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", err
	}
	published, err := os.MkdirTemp(cache, ".publish-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(published)
	contents, err := os.ReadFile(filepath.Join(temporary, "unit.o"))
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(published, "unit.o"), contents, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(published, entry); err != nil {
		if info, statErr := os.Stat(object); statErr != nil || !info.Mode().IsRegular() {
			return "", err
		}
	}
	return object, nil
}

func buildUnits(source, output string, options Options) error {
	return buildUnitsWithLibrary(source, output, options, "", nil)
}

func buildUnitsWithLibrary(source, output string, options Options, library string, extraLinkFlags []string) error {
	header, units, err := splitC(source)
	if err != nil {
		return err
	}
	directory, err := os.MkdirTemp("", "adamic-units-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	if library == "" {
		library, err = RuntimeLibrary("", options)
		if err != nil {
			return err
		}
	}
	compiler, err := exec.LookPath("clang")
	if err != nil {
		return err
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		return err
	}
	runtimeFiles, err := readRuntime(runtime, "runtime")
	if err != nil {
		return err
	}
	common := []runtimeFile{{"units.h", []byte(header)}}
	for _, file := range runtimeFiles {
		if strings.HasSuffix(file.name, ".h") {
			common = append(common, file)
		}
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	cache = filepath.Join(cache, "adamic", "units")
	jobs := options.Jobs
	if jobs == 0 {
		if value := os.Getenv("ADAMIC_NATIVE_JOBS"); value != "" {
			jobs, err = strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("native: invalid ADAMIC_NATIVE_JOBS: %w", err)
			}
		}
	}
	if jobs == 0 {
		jobs = 1
	}
	if jobs < 1 {
		return fmt.Errorf("native: split jobs must be positive")
	}
	if jobs > len(units) {
		jobs = len(units)
	}
	objects := make([]string, len(units))
	errors := make([]error, len(units))
	work := make(chan int)
	var workers sync.WaitGroup
	for j := 0; j < jobs; j++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range work {
				unit := units[index]
				files := append([]runtimeFile{{unit.name, []byte(unit.source)}}, common...)
				objects[index], errors[index] = compileUnit(unit, files, Flags(options), compiler, string(version), cache, directory, os.Getenv("ADAMIC_GATE_UNCACHED") == "1")
			}
		}()
	}
	for index := range units {
		work <- index
	}
	close(work)
	workers.Wait()
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	arguments := append(Flags(options), "-o", output)
	arguments = append(arguments, objects...)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, extraLinkFlags...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
		return fmt.Errorf("native: linking units: %w\n%s", err, output)
	}
	return nil
}
