package native

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
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
	for index := 0; index < len(source); {
		start := index
		character := source[index]
		if strings.ContainsRune(" \t\r\n", rune(character)) {
			index++
			continue
		}
		if character == '#' {
			for index < len(source) && source[index] != '\n' {
				index++
			}
			tokens = append(tokens, cToken{source[start:index], start, index})
			continue
		}
		if character == '/' && index+1 < len(source) && source[index+1] == '/' {
			for index < len(source) && source[index] != '\n' {
				index++
			}
			continue
		}
		if character == '/' && index+1 < len(source) && source[index+1] == '*' {
			end := strings.Index(source[index+2:], "*/")
			if end < 0 {
				return nil, fmt.Errorf("native: split: unterminated comment")
			}
			index += end + 4
			continue
		}
		if character == '"' || character == '\'' {
			index++
			for index < len(source) && source[index] != character {
				if source[index] == '\\' {
					index++
				}
				index++
			}
			if index >= len(source) {
				return nil, fmt.Errorf("native: split: unterminated literal")
			}
			index++
		} else if identifierByte(character) {
			for index < len(source) && identifierByte(source[index]) {
				index++
			}
		} else {
			index++
		}
		tokens = append(tokens, cToken{source[start:index], start, index})
	}
	return tokens, nil
}
func identifierByte(character byte) bool {
	return character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9' ||
		character == '_'
}

type cDeclaration struct {
	tokens            []cToken
	body, initializer int
	function          bool
	name              string
}

func splitDeclarations(source string) ([]cDeclaration, error) {
	tokens, cTokensError := cTokens(source)
	if cTokensError != nil {
		return nil, cTokensError
	}
	var declarations []cDeclaration
	for begin := 0; begin < len(tokens); {
		if strings.HasPrefix(tokens[begin].text, "#") {
			declarations = append(
				declarations,
				cDeclaration{tokens: tokens[begin : begin+1], body: -1, initializer: -1},
			)
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
		// Types have no separately owned storage. Keep them out of the symbol
		// rewrite table and copy them once into the shared header, in source order.
		if generatedTypeDeclaration(declaration.tokens) {
			declarations = append(declarations, declaration)
			begin = end
			continue
		}
		limit := len(declaration.tokens)
		if declaration.body >= 0 {
			limit = declaration.body
		}
		if declaration.initializer >= 0 {
			limit = declaration.initializer
		}
		// Generated declarators are a single identifier, followed by an array or argument list.
		// An attribute's parentheses name nothing: regexp's compiled programs are declared
		// `static const adamic_regex_program name __attribute__((aligned(64))) = {...}`.
		beforeAttribute := ""
		prototype := false
		for index := 1; index < limit; index++ {
			if declaration.tokens[index].text == "__attribute__" {
				if beforeAttribute == "" {
					beforeAttribute = declaration.tokens[index-1].text
				}
				nesting := 0
				for index+1 < limit {
					index++
					if declaration.tokens[index].text == "(" {
						nesting++
					} else if declaration.tokens[index].text == ")" {
						nesting--
						if nesting == 0 {
							break
						}
					}
				}
				continue
			}
			if declaration.tokens[index].text == "(" {
				prototype = true
				declaration.name = declaration.tokens[index-1].text
				break
			}
			if declaration.tokens[index].text == "[" {
				declaration.name = declaration.tokens[index-1].text
				break
			}
		}
		if declaration.name == "" && beforeAttribute != "" {
			declaration.name = beforeAttribute
		}
		if declaration.name == "" && limit >= 2 {
			declaration.name = declaration.tokens[limit-1].text
		}
		if declaration.name == ";" && limit >= 3 {
			declaration.name = declaration.tokens[limit-2].text
		}
		// Runtime entry points may be declared by an emitted helper's consumer.
		// Their prototypes belong in every unit's header, with external linkage and
		// their API names intact. Non-static definitions remain unsupported.
		if prototype &&
			!declaration.function &&
			declaration.initializer < 0 &&
			declaration.tokens[0].text != "static" &&
			cName.MatchString(declaration.name) {
			declaration.name = ""
			declarations = append(declarations, declaration)
			begin = end
			continue
		}
		if declaration.name != "main" &&
			(!cName.MatchString(declaration.name) || declaration.tokens[0].text != "static") {
			return nil, fmt.Errorf("native: split: unsupported declaration near %q", declaration.tokens[0].text)
		}
		declarations = append(declarations, declaration)
		begin = end
	}
	return declarations, nil
}

// generatedTypeDeclaration recognizes the emitter's object type declarations,
// including array fields inside anonymous frame typedefs. A struct followed by
// a variable declarator is storage and must still pass the static-symbol rules.
func generatedTypeDeclaration(tokens []cToken) bool {
	if len(tokens) < 3 || tokens[len(tokens)-1].text != ";" {
		return false
	}
	generated := func(name string) bool { return strings.HasPrefix(name, "adamic_") && cName.MatchString(name) }
	if tokens[0].text == "typedef" {
		return len(tokens) >= 4 && generated(tokens[len(tokens)-2].text)
	}
	if tokens[0].text == "struct" && generated(tokens[1].text) {
		return len(tokens) == 3 || tokens[2].text == "{" && tokens[len(tokens)-2].text == "}"
	}
	return false
}

// splitC gives generated static symbols external definitions, except inline helpers whose
// definitions stay in the shared header with static linkage. Token replacement never
// touches literal bytes. The fixed namespace cannot collide with runtime API names. Existing
// program-wide indexes distinguish symbols; no whole-program content hash invalidates all units.
func splitC(source string) (string, []compilationUnit, error) {
	declarations, splitDeclarationsError := splitDeclarations(source)
	if splitDeclarationsError != nil {
		return "", nil, splitDeclarationsError
	}
	names := map[string]string{}
	for _, declaration := range declarations {
		if declaration.name != "" && declaration.name != "main" {
			names[declaration.name] = "adamic_unit_" + declaration.name
		}
	}
	rewrite := func(tokens []cToken) string {
		if len(tokens) == 0 {
			return ""
		}
		var out strings.Builder
		position := tokens[0].start
		for _, token := range tokens {
			out.WriteString(source[position:token.start])
			name, found := names[token.text]
			if found {
				out.WriteString(name)
			} else {
				out.WriteString(token.text)
			}
			position = token.end
		}
		return out.String()
	}
	var header, state strings.Builder
	header.WriteString("#ifndef ADAMIC_UNITS_H\n#define ADAMIC_UNITS_H\n")
	var units []compilationUnit
	defined := map[string]bool{}
	for _, declaration := range declarations {
		if declaration.name == "" {
			header.WriteString(rewrite(declaration.tokens) + "\n")
			continue
		}
		// Every using translation unit needs the definition of a static inline helper.
		if declaration.function && declaration.tokens[0].text == "static" && declaration.tokens[1].text == "inline" {
			header.WriteString(rewrite(declaration.tokens) + "\n")
			continue
		}
		tokens := declaration.tokens
		if tokens[0].text == "static" {
			tokens = tokens[1:]
		}
		if declaration.function {
			// body indexes still refer to the original tokens, before static was removed.
			signature := rewrite(declaration.tokens[1:declaration.body])
			if declaration.name == "main" {
				signature = rewrite(declaration.tokens[:declaration.body])
			}
			header.WriteString(signature + ";\n")
			units = append(units, compilationUnit{declaration.name + ".c", rewrite(tokens) + "\n"})
			continue
		}
		limit := len(declaration.tokens) - 1
		if declaration.initializer >= 0 {
			limit = declaration.initializer
		}
		header.WriteString("extern " + rewrite(declaration.tokens[1:limit]) + ";\n")
		// Class forward declarations precede their initialized definition. A slot cache without an
		// initializer is a real zero-initialized definition, so preserve it in the state unit.
		if declaration.initializer < 0 && strings.HasPrefix(declaration.name, "adamic_class_") {
			continue
		}
		if defined[declaration.name] {
			return "", nil, fmt.Errorf("native: split: duplicate definition %s", declaration.name)
		}
		defined[declaration.name] = true
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
	for index := range units {
		units[index].source = unmarked(
			resolveSourceLines("#include \"units.h\"\n"+units[index].source, "/adamic-units/"+units[index].name),
		)
	}
	return unmarked(header.String()), units, nil
}

var unitBuilds sync.Map

// unitFirst is the line every split unit begins with (splitC), the probe its build's headers are read by.
const unitFirst = "#include \"units.h\"\n"

// unitCheckVariable sets how often a unit taken from the cache is compiled again to check its bytes: "every" checks
// every hit; unset, one in unitCheckOneIn. Any other value checks every hit.
const unitCheckVariable = "ADAMIC_NATIVE_UNIT_CHECK"

// unitCheckOneIn is how many hits share one cold check, unless unitCheckVariable asks for every one.
const unitCheckOneIn = 20

// unitCacheOffVariable, set to "off", has every unit compile uncached, as a house-wide switch can set it.
const unitCacheOffVariable = "ADAMIC_NATIVE_UNIT_CACHE"

// unitCacheOffName is the mark a mismatched cold check leaves in the unit cache: while it stands where this process
// writes or under the build cache, every unit compiles uncached, until a hand reads it and removes it.
const unitCacheOffName = "off.json"

// unitChecked says whether this hit is compiled again to check it (unitCheckVariable).
func unitChecked() bool {
	value := os.Getenv(unitCheckVariable)
	if value == "" {
		return rand.IntN(unitCheckOneIn) == 0
	}
	return true
}

// unitCacheOff says why the unit cache is off here, when it is: unitCacheOffVariable, or a mismatch's mark standing in
// cache (unitCacheOffName).
func unitCacheOff(cache string) (string, bool) {
	if os.Getenv(unitCacheOffVariable) == "off" {
		return unitCacheOffVariable + "=off", true
	}
	if mark, ok := cachedFile(filepath.Join(cache, unitCacheOffName)); ok {
		contents, _ := os.ReadFile(mark)
		return fmt.Sprintf("a cold check's mismatch is marked at %s: %s", mark, bytes.TrimSpace(contents)), true
	}
	return "", false
}

// A unitMismatch is a cached unit whose cold compile made other bytes: a key part missing, a hole, or a compile that
// isn't deterministic. Any of them means the cache isn't trusted until it's read.
type unitMismatch struct {
	Unit   string    `json:"unit"`
	Key    string    `json:"key"`
	Cached string    `json:"cached"`
	Cold   string    `json:"cold"`
	At     time.Time `json:"at"`
}

// UnitMismatchPrefix opens the one line a cold check's mismatch prints, which Loom's builder reads to turn the unit
// cache off for every builder (#m62spx0, Node's M1).
const UnitMismatchPrefix = "native: unit cache mismatch: "

// unitMismatchLine is mismatch as that line says it: the unit, its key and both digests.
func unitMismatchLine(mismatch unitMismatch) string {
	return fmt.Sprintf(
		"%sunit %s key %s cached %s cold %s",
		UnitMismatchPrefix,
		mismatch.Unit,
		mismatch.Key,
		mismatch.Cached,
		mismatch.Cold,
	)
}

// unitBuildLog is the file a build's census lines go to, where Loom's builder reads each product's (adamic's
// internal/buildcache note and record). A passing product test's own output is thrown away, so a mismatch's line goes
// here too, or no builder hears of it (Node's M1).
const unitBuildLog = "ADAMIC_BUILD_LOG"

// noteUnitMismatch appends mismatch's line to the build log when one is named, in one write a line long.
func noteUnitMismatch(mismatch unitMismatch) {
	path := os.Getenv(unitBuildLog)
	if path == "" {
		return
	}
	file, openError := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if openError != nil {
		fmt.Fprintf(os.Stderr, "native: the unit cache mismatch couldn't be noted in %s: %v\n", path, openError)
		return
	}
	defer file.Close()
	if _, writeStringError := file.WriteString(unitMismatchLine(mismatch) + "\n"); writeStringError != nil {
		fmt.Fprintf(os.Stderr, "native: the unit cache mismatch couldn't be noted in %s: %v\n", path, writeStringError)
	}
}

// markUnitCacheOff leaves mismatch's mark in cache, made whole before it's in place.
func markUnitCacheOff(cache string, mismatch unitMismatch) error {
	content, encodeError := json.Marshal(mismatch)
	if encodeError != nil {
		return encodeError
	}
	if createDirectoryError := os.MkdirAll(cache, 0o755); createDirectoryError != nil {
		return createDirectoryError
	}
	temporary, temporaryFileError := os.CreateTemp(cache, ".off-")
	if temporaryFileError != nil {
		return temporaryFileError
	}
	defer os.Remove(temporary.Name())
	if _, writeError := temporary.Write(append(content, '\n')); writeError != nil {
		temporary.Close()
		return writeError
	}
	if closeError := temporary.Close(); closeError != nil {
		return closeError
	}
	return os.Rename(temporary.Name(), filepath.Join(cache, unitCacheOffName))
}

// compileUnit is unit's object: from cache by its key under keying, or compiled from its .c in a directory of its own
// under directory holding common beside it, and kept in cache. A nil keying, or a unit whose source the key can't take
// (unitSourceRefusal), compiles uncached. A sampled hit is compiled again, and a mismatch marks the cache off, prints
// UnitMismatchPrefix's line naming the unit, the key and both digests, and is the cold object, never an error.
func compileUnit(
	unit compilationUnit,
	common []runtimeFile,
	flags []string,
	compiler string,
	keying *unitKeying,
	cache, directory string,
	jobserver *Jobserver,
	sanitized bool,
) (string, error) {
	return compileUnitAs(unit, common, flags, compiler, keying, cache, directory, jobserver, "off", sanitized)
}

// compileUnitAs is compileUnit with uncached, the outcome its build log line says when the unit compiles uncached: off
// for a build that read no cache, repass for compileUnits' second pass after a mismatch (Performance on #m62spx0), so
// a reader counting units or a hit rate leaves the pass out from the line alone, in a log several processes append to,
// whose order no reader can trust.
func compileUnitAs(
	unit compilationUnit,
	common []runtimeFile,
	flags []string,
	compiler string,
	keying *unitKeying,
	cache, directory string,
	jobserver *Jobserver,
	uncached string,
	sanitized bool,
) (string, error) {
	environment := toolchain.Env()
	if keying != nil {
		environment = keying.environment
		if why := unitSourceRefusal(unit, keying.first); why != "" {
			fmt.Fprintf(os.Stderr, "native: unit %s compiles uncached: %s\n", unit.name, why)
			keying = nil
		}
	}
	if keying == nil {
		object, clang, compileUnitObjectError := compileUnitObject(
			unit,
			common,
			flags,
			compiler,
			environment,
			directory,
			jobserver,
			sanitized,
		)
		if compileUnitObjectError != nil {
			return "", compileUnitObjectError
		}
		// An uncached unit is never keyed, since keying is what reads clang's facts before it runs, so its line names
		// no key.
		buildLog("cCompile", unit.name, "", uncached, clang)
		path := filepath.Join(directory, "uncached-"+unit.name+".o")
		return path, os.WriteFile(path, object, 0o644)
	}
	// keyed is the seconds the unit's key took, a hit's whole cost here, since a hit runs no clang (#m62spx0).
	var keyed time.Duration
	var key string
	timed(&keyed, func() error { key = keying.key(unit); return nil })
	entry := filepath.Join(cache, key)
	object := filepath.Join(entry, "unit.o")
	value, _ := unitBuilds.LoadOrStore(entry, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	// Where this process writes, then the build cache: the link the units' directory makes to it names wherever it was
	// found.
	if found, ok := cachedFile(object); ok {
		if !keying.checked(unit) {
			buildLog("cCompile", unit.name, key, "hit", keyed)
			return found, nil
		}
		cached, readError := os.ReadFile(found)
		if readError != nil {
			return "", readError
		}
		cold, clang, compileUnitObjectError := compileUnitObject(
			unit,
			common,
			flags,
			compiler,
			environment,
			directory,
			jobserver,
			sanitized,
		)
		if compileUnitObjectError != nil {
			return "", compileUnitObjectError
		}
		if !bytes.Equal(cached, cold) {
			mismatch := unitMismatch{
				Unit:   unit.name,
				Key:    key,
				Cached: fmt.Sprintf("%x", sha256.Sum256(cached)),
				Cold:   fmt.Sprintf("%x", sha256.Sum256(cold)),
				At:     time.Now().UTC(),
			}
			// The cold bytes were compiled fresh from this unit's sources, so they're its object: the build goes on
			// with them, green, and the hole is the house's to read, never a red charged to whatever change was
			// building.
			marked := "marked off at " + filepath.Join(cache, unitCacheOffName)
			if markUnitCacheOffError := markUnitCacheOff(cache, mismatch); markUnitCacheOffError != nil {
				marked = "its mark couldn't be written: " + markUnitCacheOffError.Error()
			}
			fmt.Fprintln(os.Stderr, unitMismatchLine(mismatch)+"; the unit cache is "+marked)
			noteUnitMismatch(mismatch)
			keying.mismatched.Store(true)
			// The cold bytes are the object, so clang made it: a miss, with the cold compile's seconds.
			buildLog("cCompile", unit.name, key, "miss", keyed+clang)
			path := filepath.Join(directory, "checked-"+unit.name+".o")
			return path, os.WriteFile(path, cold, 0o644)
		}
		// The cache's object is used, so it's a hit, its seconds the key and the check's compile.
		buildLog("cCompile", unit.name, key, "hit", keyed+clang)
		return found, nil
	}
	contents, clang, compileUnitObjectError := compileUnitObject(
		unit,
		common,
		flags,
		compiler,
		environment,
		directory,
		jobserver,
		sanitized,
	)
	if compileUnitObjectError != nil {
		return "", compileUnitObjectError
	}
	// clang compiled it, so it is a miss even when another process published the same entry first.
	buildLog("cCompile", unit.name, key, "miss", keyed+clang)
	if createDirectoryError := os.MkdirAll(cache, 0o755); createDirectoryError != nil {
		return "", createDirectoryError
	}
	published, temporaryDirectoryError := os.MkdirTemp(cache, ".publish-")
	if temporaryDirectoryError != nil {
		return "", temporaryDirectoryError
	}
	defer os.RemoveAll(published)
	if writeError := os.WriteFile(filepath.Join(published, "unit.o"), contents, 0o644); writeError != nil {
		return "", writeError
	}
	if renameError := os.Rename(published, entry); renameError != nil {
		found, ok := cachedFile(object)
		if !ok {
			return "", renameError
		}
		return found, nil
	}
	return object, nil
}

// compileUnitObject compiles unit from its .c, in a directory of its own under directory holding common beside it,
// under environment exactly, and is the object's bytes and how long clang ran, timed inside its jobserver.Hold so a
// token's wait isn't counted (library.go's buildLog).
func compileUnitObject(
	unit compilationUnit,
	common []runtimeFile,
	flags []string,
	compiler string,
	environment []string,
	directory string,
	jobserver *Jobserver,
	sanitized bool,
) ([]byte, time.Duration, error) {
	temporary, temporaryDirectoryError := os.MkdirTemp(directory, ".unit-")
	if temporaryDirectoryError != nil {
		return nil, 0, temporaryDirectoryError
	}
	defer os.RemoveAll(temporary)
	for _, file := range append([]runtimeFile{{unit.name, []byte(unit.source)}}, common...) {
		if writeError := os.WriteFile(filepath.Join(temporary, file.name), file.contents, 0o644); writeError != nil {
			return nil, 0, writeError
		}
	}
	command := clangIn(
		temporary,
		compiler,
		append(slices.Clone(flags), unitCompileArguments(unit.name, temporary, flags, sanitized)...)...,
	)
	// The environment the key read, with the directory CommandIn names as the compile's own.
	working := []string{}
	for _, setting := range command.Env {
		if strings.HasPrefix(setting, "PWD=") {
			working = append(working, setting)
		}
	}
	command.Env = append(slices.Clone(environment), working...)
	var output []byte
	var clang time.Duration
	if holdError := jobserver.Hold(func() error {
		return timed(&clang, func() (err error) { output, err = command.CombinedOutput(); return err })
	}); holdError != nil {
		return nil, 0, fmt.Errorf("native: compiling unit %s: %w\n%s", unit.name, holdError, output)
	}
	object, readError := os.ReadFile(filepath.Join(temporary, "unit.o"))
	return object, clang, readError
}

// compileUnits is every unit's object, jobs at once (compileUnit). Once any unit's cold check mismatches, the build's
// cache is proved incomplete, not that unit's alone: every unit compiles again uncached, the hits already taken
// included, so nothing is linked from keys this build caught short (Node's M4).
func compileUnits(
	units []compilationUnit,
	common []runtimeFile,
	flags []string,
	compiler string,
	keying *unitKeying,
	cache, directory string,
	jobs int,
	jobserver *Jobserver,
	sanitized bool,
) ([]string, error) {
	objects := make([]string, len(units))
	failures := make([]error, len(units))
	pass := func(keying *unitKeying, uncached string) error {
		work := make(chan int)
		var workers sync.WaitGroup
		for range jobs {
			workers.Go(func() {
				for index := range work {
					objects[index], failures[index] = compileUnitAs(
						units[index],
						common,
						flags,
						compiler,
						keying,
						cache,
						directory,
						jobserver,
						uncached,
						sanitized,
					)
				}
			})
		}
		for index := range units {
			work <- index
		}
		close(work)
		workers.Wait()
		return errors.Join(failures...)
	}
	if passError := pass(keying, "off"); passError != nil {
		return nil, passError
	}
	if keying != nil && keying.mismatched.Load() {
		fmt.Fprintln(
			os.Stderr,
			"native: units compile uncached: a unit's cold check mismatched in this build, so every unit compiles again",
		)
		if passError := pass(nil, "repass"); passError != nil {
			return nil, passError
		}
	}
	return objects, nil
}

func buildUnits(source, output string, options Options) error {
	return buildUnitsWithLibrary(source, output, options, "", "")
}

// buildUnitsWithLibrary links the units with library, or the program's own runtime when it is empty, and with the
// checker archive at the absolute path archive (checkerArchivePath) when it isn't empty.
func buildUnitsWithLibrary(source, output string, options Options, library string, archive string) error {
	header, units, splitCError := splitC(source)
	if splitCError != nil {
		return splitCError
	}
	jobserver, defaultJobserverError := defaultJobserver()
	if defaultJobserverError != nil {
		return defaultJobserverError
	}
	directory, temporaryDirectoryError := os.MkdirTemp("", "adamic-units-")
	if temporaryDirectoryError != nil {
		return temporaryDirectoryError
	}
	defer os.RemoveAll(directory)
	if library == "" {
		// Split and Jobs describe program compilation, never a different runtime archive.
		runtimeOptions := options
		runtimeOptions.Split, runtimeOptions.Jobs = false, 0
		var runtimeLibraryForSourceError error
		library, runtimeLibraryForSourceError = RuntimeLibraryForSource("", source, runtimeOptions)
		if runtimeLibraryForSourceError != nil {
			return runtimeLibraryForSourceError
		}
	}
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		return pathError
	}
	runtimeFiles, readRuntimeError := readRuntime(runtime, "runtime")
	if readRuntimeError != nil {
		return readRuntimeError
	}
	common := []runtimeFile{{"units.h", []byte(header)}}
	for _, file := range runtimeFiles {
		if strings.HasSuffix(file.name, ".h") {
			common = append(common, file)
		}
	}
	cache, cacheDirectoryError := cacheDirectory("native-units")
	if cacheDirectoryError != nil {
		return cacheDirectoryError
	}
	jobs := options.Jobs
	if jobs == 0 {
		if value := os.Getenv("ADAMIC_NATIVE_JOBS"); value != "" {
			var parseError error
			jobs, parseError = strconv.Atoi(value)
			if parseError != nil {
				return fmt.Errorf("native: invalid ADAMIC_NATIVE_JOBS: %w", parseError)
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
	// A unit includes units.h, and through it adamic.h, before any of the program's own #defines
	// (the checker build puts tsgo_runtime.h first), so its features come as flags.
	flags := append(sourceFlags(source, options), programDebugFlags(options)...)
	sanitized := options.Sanitize || options.ThreadSanitize
	// The units' key is read once, before any of them compiles: a build the key can't account for, or one the cache is
	// off for, compiles every unit uncached and says why.
	var keying *unitKeying
	if os.Getenv("ADAMIC_GATE_UNCACHED") != "1" {
		var newUnitKeyingError error
		if why, off := unitCacheOff(cache); off {
			fmt.Fprintf(os.Stderr, "native: units compile uncached: %s\n", why)
		} else if keying, newUnitKeyingError = newUnitKeying(
			unitFirst,
			common,
			flags,
			compiler,
			toolchain.Env(),
			sanitized,
		); errors.Is(newUnitKeyingError, errorUncachedUnit) {
			fmt.Fprintf(os.Stderr, "%v\n", newUnitKeyingError)
			keying = nil
		} else if newUnitKeyingError != nil {
			return newUnitKeyingError
		}
	}
	objects, compileUnitsError := compileUnits(
		units,
		common,
		flags,
		compiler,
		keying,
		cache,
		directory,
		jobs,
		jobserver,
		sanitized,
	)
	if compileUnitsError != nil {
		return compileUnitsError
	}
	// The link runs from the units' directory and gives it no input by its place on the machine: each unit's object,
	// the runtime library and the checker archive go by a link's name there, and only the output by its absolute path,
	// as BuildTSGo links. On macOS ld's debug map records every object it links that carries debug info, a sanitized
	// build's all do, by the path it was given made absolute, which -oso_prefix takes the directory back off. Named by
	// their places, a sanitized split binary held the build cache's path 57 times (measured on a Mac, Oct 11).
	output, absolutePathError := filepath.Abs(output)
	if absolutePathError != nil {
		return absolutePathError
	}
	arguments := append(LinkFlags(options), "-o", output)
	for index, object := range objects {
		name := fmt.Sprintf("unit-%d.o", index)
		if symlinkError := os.Symlink(object, filepath.Join(directory, name)); symlinkError != nil {
			return symlinkError
		}
		arguments = append(arguments, name)
	}
	if symlinkError := os.Symlink(filepath.Dir(library), filepath.Join(directory, "runtime")); symlinkError != nil {
		return symlinkError
	}
	arguments = append(arguments, RuntimeLinkFlags(filepath.Join("runtime", filepath.Base(library)))...)
	if archive != "" {
		if symlinkError := os.Symlink(archive, filepath.Join(directory, checkerArchive)); symlinkError != nil {
			return symlinkError
		}
		arguments = append(arguments, checkerArchive, "-lpthread", "-ldl")
	}
	compilerLink, regexpCompilerLinkFlagsError := regexpCompilerLinkFlags(source, options, directory)
	if regexpCompilerLinkFlagsError != nil {
		return regexpCompilerLinkFlagsError
	}
	arguments = append(arguments, compilerLink...)
	arguments = append(arguments, objectPrefixFlags()...)
	arguments = append(arguments, "-lm")
	link := clangIn(directory, compiler, arguments...)
	var linked []byte
	var linking time.Duration
	if holdError := jobserver.Hold(func() error {
		return timed(&linking, func() (err error) { linked, err = link.CombinedOutput(); return err })
	}); holdError != nil {
		return fmt.Errorf("native: linking units: %w\n%s", holdError, linked)
	}
	buildLog("cLink", filepath.Base(output), "", "off", linking)
	return nil
}
