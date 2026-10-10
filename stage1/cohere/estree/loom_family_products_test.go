package estree

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Four workers use the gate's four CPUs; source, sanitizer flags and cases stay fixed.
func estreeFamilyNativeOptions() native.Options {
	return native.Options{Sanitize: true, Split: true, Jobs: 4}
}

func TestProduct_DeepMutantsNative0(t *testing.T) {
	t.Parallel()
	started := time.Now()
	deepMutantsNative(t, deepMutantsEnumeration()[0])
	t.Logf("product: %.3fs", time.Since(started).Seconds())
}

func TestProduct_SyntaxMutantNative_000(t *testing.T) {
	t.Parallel()
	syntaxMutantNativeProduct(t, syntaxMutantEnumeration()[0])
}

func TestProduct_SyntaxMutantsSetup_000(t *testing.T) {
	t.Parallel()
	syntaxMutantPreparedProduct(t, 0)
}
func TestProduct_SyntaxMutantsSetup_001(t *testing.T) {
	t.Parallel()
	syntaxMutantPreparedProduct(t, 1)
}
func TestProduct_SyntaxMutantsSetup_002(t *testing.T) {
	t.Parallel()
	syntaxMutantPreparedProduct(t, 2)
}

func unattachedDecoratorMutation() syntaxMutant {
	return syntaxMutant{name: "unattached-decorator", file: "pipeline.ts", from: `parser.nodes[id]?.kind === 'Decorator'`, to: `parser.nodes[id]?.kind === 'UnusedDecoratorControl'`}
}

var unattachedDecoratorPrepared struct {
	once                 sync.Once
	main, binary, oracle string
}

func unattachedDecoratorReady(t *testing.T) (string, string, string) {
	t.Helper()
	unattachedDecoratorPrepared.once.Do(func() {
		unattachedDecoratorPrepared.main, unattachedDecoratorPrepared.binary = syntaxMutantProducts(t, unattachedDecoratorMutation())
		unattachedDecoratorPrepared.oracle = filepath.Join(syntaxMutantOracleProduct(t), "oracle")
	})
	if unattachedDecoratorPrepared.oracle == "" {
		t.Fatal("decorator preparation failed")
	}
	return unattachedDecoratorPrepared.main, unattachedDecoratorPrepared.binary, unattachedDecoratorPrepared.oracle
}

func TestProduct_UnattachedDecoratorLowered(t *testing.T) {
	t.Parallel()
	syntaxMutantLoweredProduct(t, unattachedDecoratorMutation())
}
func TestProduct_UnattachedDecoratorNative(t *testing.T) {
	t.Parallel()
	syntaxMutantNativeProduct(t, unattachedDecoratorMutation())
}

func TestUnattachedDecoratorControl(t *testing.T) {
	t.Parallel()
	setup := time.Now()
	main, binary, oracle := unattachedDecoratorReady(t)
	script := mutantEmittedProduct(t, main)
	t.Logf("setup: %.3fs", time.Since(setup).Seconds())
	started := time.Now()
	defer func() { t.Logf("own work: %.3fs", time.Since(started).Seconds()) }()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	list := manifest(t, []string{"@dec\nawait 1", "@dec\nx"})
	statuses := string(threePortExecute(t, ctx, oracle, "--audit", list, t.TempDir()))
	if strings.Count(statuses, `"status":"error"`) != 2 {
		t.Fatalf("Go decorator refusals changed: %s", statuses)
	}
	for name, got := range map[string][]byte{
		"Node":    threePortExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, "--manifest", list),
		"native":  threePortExecute(t, ctx, binary, "--manifest", list),
		"emitted": mutantEmittedOutput(t, main, script, "--manifest", list),
	} {
		if strings.Count(string(got), "0 Program ") != 2 {
			t.Fatal(name + " decorator control did not finish with two incorrect acceptances")
		}
		t.Log(name + ": disabled orphan guard accepts both Go-refused files; acceptance check catches it")
	}
}

// Checkpoint the checked IR so the native emitter has its own cached build stage.
// Gob retains the concrete expression/statement types; derived private facts are
// recomputed by the backend. Source snapshots remain stable for Node execution.
var estreeFamilyGobOnce sync.Once

func estreeFamilyRegisterIR() {
	estreeFamilyGobOnce.Do(func() {
		gob.Register(ir.AllocateEnvironment{})
		gob.Register(ir.ArrayConcat{})
		gob.Register(ir.ArrayFill{})
		gob.Register(ir.ArrayFrom{})
		gob.Register(ir.ArrayIndex{})
		gob.Register(ir.ArrayIsArray{})
		gob.Register(ir.ArrayJoin{})
		gob.Register(ir.ArrayLiteral{})
		gob.Register(ir.ArrayMap{})
		gob.Register(ir.ArrayPop{})
		gob.Register(ir.ArrayPush{})
		gob.Register(ir.ArrayReduce{})
		gob.Register(ir.ArrayReverse{})
		gob.Register(ir.ArraySearch{})
		gob.Register(ir.ArraySlice{})
		gob.Register(ir.ArraySort{})
		gob.Register(ir.ArraySplice{})
		gob.Register(ir.ArrayVisit{})
		gob.Register(ir.Assign{})
		gob.Register(ir.Binary{})
		gob.Register(ir.Block{})
		gob.Register(ir.BooleanConstant{})
		gob.Register(ir.BooleanToString{})
		gob.Register(ir.Box{})
		gob.Register(ir.Break{})
		gob.Register(ir.Call{})
		gob.Register(ir.CallClosure{})
		gob.Register(ir.CharCodeAt{})
		gob.Register(ir.CheckedCast{})
		gob.Register(ir.ClosureSelf{})
		gob.Register(ir.Coalesce{})
		gob.Register(ir.CodePoints{})
		gob.Register(ir.CollectionIterator{})
		gob.Register(ir.Comma{})
		gob.Register(ir.Concat{})
		gob.Register(ir.Conditional{})
		gob.Register(ir.Continue{})
		gob.Register(ir.Debugger{})
		gob.Register(ir.Declare{})
		gob.Register(ir.Defined{})
		gob.Register(ir.DynamicProperty{})
		gob.Register(ir.Effects{})
		gob.Register(ir.Evaluate{})
		gob.Register(ir.FileStatus{})
		gob.Register(ir.ForOf{})
		gob.Register(ir.HasAccessor{})
		gob.Register(ir.HasOwn{})
		gob.Register(ir.HasProperty{})
		gob.Register(ir.If{})
		gob.Register(ir.InstanceOf{})
		gob.Register(ir.IsNull{})
		gob.Register(ir.IsUndefined{})
		gob.Register(ir.JSONNull{})
		gob.Register(ir.JSONStringify{})
		gob.Register(ir.Labeled{})
		gob.Register(ir.Length{})
		gob.Register(ir.LibraryGlobal{})
		gob.Register(ir.Logical{})
		gob.Register(ir.LogicalAssignment{})
		gob.Register(ir.Loop{})
		gob.Register(ir.MakeClosure{})
		gob.Register(ir.MakeError{})
		gob.Register(ir.MapClear{})
		gob.Register(ir.MapDelete{})
		gob.Register(ir.MapEntries{})
		gob.Register(ir.MapForEach{})
		gob.Register(ir.MapGet{})
		gob.Register(ir.MapHas{})
		gob.Register(ir.MapKeys{})
		gob.Register(ir.MapNew{})
		gob.Register(ir.MapSet{})
		gob.Register(ir.MapSize{})
		gob.Register(ir.MapValues{})
		gob.Register(ir.MathCall{})
		gob.Register(ir.MaybeOf{})
		gob.Register(ir.MaybeToString{})
		gob.Register(ir.Narrow{})
		gob.Register(ir.NodeBufferCall{})
		gob.Register(ir.NodeFSFile{})
		gob.Register(ir.Null{})
		gob.Register(ir.NumberCall{})
		gob.Register(ir.NumberConstant{})
		gob.Register(ir.NumberFormat{})
		gob.Register(ir.NumberToString{})
		gob.Register(ir.ObjectCall{})
		gob.Register(ir.ObjectKeys{})
		gob.Register(ir.ObjectLiteral{})
		gob.Register(ir.Panic{})
		gob.Register(ir.ProgramArguments{})
		gob.Register(ir.Property{})
		gob.Register(ir.Read{})
		gob.Register(ir.ReadDirectory{})
		gob.Register(ir.ReadTextFile{})
		gob.Register(ir.RegExpCall{})
		gob.Register(ir.RegExpGroup{})
		gob.Register(ir.RegExpNew{})
		gob.Register(ir.RegExpProperty{})
		gob.Register(ir.Return{})
		gob.Register(ir.SetAdd{})
		gob.Register(ir.SetIndex{})
		gob.Register(ir.SetNew{})
		gob.Register(ir.SetProperty{})
		gob.Register(ir.SetValues{})
		gob.Register(ir.StringCall{})
		gob.Register(ir.StringConstant{})
		gob.Register(ir.StringFromCodes{})
		gob.Register(ir.StringIndex{})
		gob.Register(ir.StringLength{})
		gob.Register(ir.Switch{})
		gob.Register(ir.Throw{})
		gob.Register(ir.ToFixed{})
		gob.Register(ir.Trim{})
		gob.Register(ir.Truthy{})
		gob.Register(ir.Try{})
		gob.Register(ir.TypeOf{})
		gob.Register(ir.TypedArrayFill{})
		gob.Register(ir.TypedArrayNew{})
		gob.Register(ir.TypedArraySet{})
		gob.Register(ir.TypedArraySubarray{})
		gob.Register(ir.Unary{})
		gob.Register(ir.Undefined{})
		gob.Register(ir.UnionToString{})
		gob.Register(ir.Unwrap{})
		gob.Register(ir.Utf8At{})
		gob.Register(ir.Utf8Length{})
		gob.Register(ir.Void{})
		gob.Register(ir.WeakOf{})
		gob.Register(ir.WeakTarget{})
		gob.Register(ir.WriteLine{})
		gob.Register(ir.WriteTextFile{})
	})
}
func estreeFamilyLowerCheckpoint(dir string, snapshot map[string][]byte) error {
	source := filepath.Join(dir, "source")
	if err := os.Mkdir(source, 0755); err != nil {
		return err
	}
	for name, data := range snapshot {
		if err := os.WriteFile(filepath.Join(source, name), data, 0644); err != nil {
			return err
		}
	}
	program, err := load.Load([]string{filepath.Join(source, "main.ts")})
	if err != nil {
		return err
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return err
	}
	estreeFamilyRegisterIR()
	file, err := os.Create(filepath.Join(dir, "program.gob"))
	if err != nil {
		return err
	}
	err = gob.NewEncoder(file).Encode(lowered)
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("encode checked IR: %w", err)
	}
	return closeErr
}
func estreeFamilyReadCheckpoint(dir string) (*ir.Program, error) {
	estreeFamilyRegisterIR()
	file, err := os.Open(filepath.Join(dir, "program.gob"))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var program ir.Program
	if err := gob.NewDecoder(file).Decode(&program); err != nil {
		return nil, err
	}
	return &program, nil
}
func estreeFamilyCopySource(from, to string) error {
	if err := os.Mkdir(filepath.Join(to, "source"), 0755); err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(from, "source", "*.ts"))
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(to, "source", filepath.Base(file)), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
func TestProduct_DeepMutantsIR0(t *testing.T) {
	t.Parallel()
	deepMutantsIRProduct(t, deepMutantsEnumeration()[0])
}
func TestProduct_DeepMutantsIR1(t *testing.T) {
	t.Parallel()
	deepMutantsIRProduct(t, deepMutantsEnumeration()[1])
}
func TestProduct_DeepMutantsIR2(t *testing.T) {
	t.Parallel()
	deepMutantsIRProduct(t, deepMutantsEnumeration()[2])
}

// Hash actual Go dependencies (including embeds and local modules), rather than
// unrelated upstream test corpora. This caches metadata only, never a build.
var estreeFamilyDependencies struct {
	once     sync.Once
	files    []string
	external []string
	err      error
}

func estreeFamilyDependencyFiles(t *testing.T) []string {
	t.Helper()
	estreeFamilyDependencies.once.Do(func() {
		repo := root(t)
		command := exec.Command("go", "list", "-deps", "-json", "./internal/load", "./internal/lower", "./internal/native", "./internal/javascript", "github.com/system-inc/cohere/internal/format/estree")
		command.Dir = repo
		output, err := command.Output()
		if err != nil {
			estreeFamilyDependencies.err = fmt.Errorf("Go dependencies: %w", err)
			return
		}
		decoder := json.NewDecoder(bytes.NewReader(output))
		files := map[string]bool{"go.mod": true, "go.work": true, "stage1/typescript": true, "stage1/cohere/estree": true}
		add := func(path string) error {
			relative, err := filepath.Rel(repo, path)
			if err != nil {
				return err
			}
			if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				estreeFamilyDependencies.external = append(estreeFamilyDependencies.external, fmt.Sprintf("external=%s:%x", path, sha256.Sum256(data)))
				return nil
			}
			files[filepath.ToSlash(relative)] = true
			return nil
		}
		for {
			var item struct {
				Standard                                                                   bool
				Dir                                                                        string
				GoFiles, CgoFiles, CFiles, CXXFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
				Module                                                                     *struct{ GoMod string }
				Error                                                                      *struct{ Err string }
			}
			if err := decoder.Decode(&item); err == io.EOF {
				break
			} else if err != nil {
				estreeFamilyDependencies.err = err
				return
			}
			if item.Error != nil {
				estreeFamilyDependencies.err = fmt.Errorf("Go dependency: %s", item.Error.Err)
				return
			}
			if item.Standard {
				continue
			}
			for _, group := range [][]string{item.GoFiles, item.CgoFiles, item.CFiles, item.CXXFiles, item.HFiles, item.SFiles, item.SysoFiles, item.EmbedFiles} {
				for _, name := range group {
					if err := add(filepath.Join(item.Dir, name)); err != nil {
						estreeFamilyDependencies.err = err
						return
					}
				}
			}
			if item.Module != nil && item.Module.GoMod != "" {
				if err := add(item.Module.GoMod); err != nil {
					estreeFamilyDependencies.err = err
					return
				}
				sum := filepath.Join(filepath.Dir(item.Module.GoMod), "go.sum")
				if _, err := os.Stat(sum); err == nil {
					if err := add(sum); err != nil {
						estreeFamilyDependencies.err = err
						return
					}
				}
			}
		}
		for file := range files {
			estreeFamilyDependencies.files = append(estreeFamilyDependencies.files, file)
		}
		sort.Strings(estreeFamilyDependencies.files)
		sort.Strings(estreeFamilyDependencies.external)
	})
	if estreeFamilyDependencies.err != nil {
		t.Fatal(estreeFamilyDependencies.err)
	}
	return estreeFamilyDependencies.files
}

func TestProduct_SyntaxMutantIR_000(t *testing.T) {
	t.Parallel()
	syntaxMutantIRProduct(t, syntaxMutantEnumeration()[0])
}
func TestProduct_SyntaxMutantIR_001(t *testing.T) {
	t.Parallel()
	syntaxMutantIRProduct(t, syntaxMutantEnumeration()[1])
}
func TestProduct_SyntaxMutantIR_002(t *testing.T) {
	t.Parallel()
	syntaxMutantIRProduct(t, syntaxMutantEnumeration()[2])
}
func TestProduct_UnattachedDecoratorIR(t *testing.T) {
	t.Parallel()
	syntaxMutantIRProduct(t, unattachedDecoratorMutation())
}

func estreeFamilyDependencyFlags(t *testing.T) []string {
	t.Helper()
	estreeFamilyDependencyFiles(t)
	return estreeFamilyDependencies.external
}
