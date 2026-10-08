package main

// Package-local scheduler experiment. It does not expose a new facts ABI or
// claim observed-read enforcement: the pinned checker has no read-event hook.
import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

var nextBatchGeneration atomic.Uint64

type batchID struct{ Owner, Generation, Local uint64 }
type batchSelector struct {
	Start, End int
	Kind       ast.Kind
}
type batchQuestion struct {
	Selector        batchSelector
	Operation       string
	Operand         batchID
	FromQuestion    bool
	OperandQuestion int
}
type plannedFile struct {
	Path      string
	Questions []batchQuestion
}
type batchAnswer struct {
	Value, Error string
	ID           batchID
}
type batchOwner struct {
	Checker            *checker.Checker
	Number, Generation uint64
	IDs                map[*checker.Type]uint64
	Types              []*checker.Type
}

func (o *batchOwner) intern(t *checker.Type) batchID {
	id, ok := o.IDs[t]
	if !ok {
		id = uint64(len(o.Types))
		o.IDs[t] = id
		o.Types = append(o.Types, t)
	}
	return batchID{o.Number, o.Generation, id}
}
func (o *batchOwner) operand(id batchID) (*checker.Type, error) {
	if id.Owner != o.Number {
		return nil, fmt.Errorf("cross-owner operand: refused; re-ask on owner %d", id.Owner)
	}
	if id.Generation != o.Generation {
		return nil, fmt.Errorf("stale generation")
	}
	if id.Local >= uint64(len(o.Types)) {
		return nil, fmt.Errorf("unknown local ID")
	}
	return o.Types[id.Local], nil
}

// Plans are complete before workers start, and copied so caller mutations cannot
// change an in-flight batch. Results retain file/question slots, not arrival order.
func freezePlan(plan []plannedFile) []plannedFile {
	out := make([]plannedFile, len(plan))
	for i, p := range plan {
		out[i] = plannedFile{p.Path, append([]batchQuestion(nil), p.Questions...)}
	}
	return out
}
func batchProgram(config string, n int) *compiler.Program {
	fs := cachedvfs.From(bundled.WrapFS(osvfs.FS()))
	host := compiler.NewCachedFSCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	parsed, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(config)), nil, nil, fs, nil)
	if parsed == nil || len(diagnostics) != 0 || len(parsed.Errors) != 0 {
		panic("invalid batch config")
	}
	parsed.CompilerOptions().Checkers = &n
	return compiler.NewProgram(compiler.ProgramOptions{Config: parsed, Host: host, SingleThreaded: core.TSFalse})
}
func resolveSelector(file *ast.SourceFile, s batchSelector) *ast.Node {
	var found *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if found != nil {
			return
		}
		if node.Pos() == s.Start && node.End() == s.End && node.Kind == s.Kind {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return found != nil })
	}
	walk(file.AsNode())
	return found
}
func (o *batchOwner) answer(file *ast.SourceFile, q batchQuestion) batchAnswer {
	switch q.Operation {
	case "type-name":
		node := resolveSelector(file, q.Selector)
		if node == nil {
			return batchAnswer{Error: "exact selector absent"}
		}
		typed := o.Checker.GetTypeAtLocation(node)
		id := o.intern(typed)
		return batchAnswer{Value: o.Checker.TypeToString(typed), ID: id}
	case "id-name":
		typed, err := o.operand(q.Operand)
		if err != nil {
			return batchAnswer{Error: err.Error()}
		}
		return batchAnswer{Value: o.Checker.TypeToString(typed), ID: q.Operand}
	default:
		return batchAnswer{Error: "unsupported scout operation: " + q.Operation}
	}
}

type batchSession struct {
	Program    *compiler.Program
	Generation uint64
	Owners     map[*checker.Checker]*batchOwner
}

func newBatchSession(p *compiler.Program) *batchSession {
	return &batchSession{Program: p, Generation: nextBatchGeneration.Add(1), Owners: map[*checker.Checker]*batchOwner{}}
}

// A session owns its checker tables for the entire program lifetime. Plans on
// the same session execute sequentially; distinct sessions own distinct programs.
func executePlan(session *batchSession, plan []plannedFile, reverseCompletion bool) [][]batchAnswer {
	p, generation := session.Program, session.Generation
	plan = freezePlan(plan)
	answers := make([][]batchAnswer, len(plan))
	owners := session.Owners
	jobs := map[*batchOwner][]int{}
	files := make([]*ast.SourceFile, len(plan))
	// Resolve affinity before any checker is used concurrently. The nonexclusive
	// accessor is safe here and later because exactly one worker owns each checker.
	for i, row := range plan {
		file := p.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(row.Path)))
		if file == nil {
			panic("planned file absent")
		}
		files[i] = file
		c, release := p.GetTypeCheckerForFile(context.Background(), file)
		release()
		owner := owners[c]
		if owner == nil {
			owner = &batchOwner{Checker: c, Number: uint64(len(owners) + 1), Generation: generation, IDs: map[*checker.Type]uint64{}}
			owners[c] = owner
		}
		jobs[owner] = append(jobs[owner], i)
		answers[i] = make([]batchAnswer, len(row.Questions))
	}
	done := make(chan struct{}, len(jobs))
	for owner, indices := range jobs {
		go func(o *batchOwner, indices []int) {
			if reverseCompletion {
				time.Sleep(time.Duration(len(owners)-int(o.Number)) * time.Millisecond)
			}
			for _, i := range indices {
				for j, q := range plan[i].Questions {
					if q.FromQuestion {
						if q.OperandQuestion < 0 || q.OperandQuestion >= j {
							answers[i][j] = batchAnswer{Error: "dependency must precede question"}
							continue
						}
						previous := answers[i][q.OperandQuestion]
						if previous.Error != "" {
							answers[i][j] = batchAnswer{Error: "operand question failed"}
							continue
						}
						q.Operand = previous.ID
					}
					answers[i][j] = o.answer(files[i], q)
				}
			}
			done <- struct{}{}
		}(owner, indices)
	}
	for range jobs {
		<-done
	}
	// A concrete ID submitted through a different file owner is first refused by
	// that worker. Once workers are drained, transfer exclusive execution to the
	// coordinator and re-ask the issuing owner. No checker is concurrently active.
	for i, row := range plan {
		for j, q := range row.Questions {
			if q.Operation != "id-name" || q.FromQuestion || answers[i][j].Error == "" {
				continue
			}
			for _, owner := range session.Owners {
				if owner.Number == q.Operand.Owner {
					answers[i][j] = owner.answer(files[i], q)
					break
				}
			}
		}
	}
	return answers
}
func replayBatch(plan []plannedFile, answers [][]batchAnswer) string {
	// Owner-local IDs are handles, not finding/replay payload; the canonical
	// transcript preserves selectors and semantic outcomes in input order.
	var out strings.Builder
	for i, file := range plan {
		out.WriteString(frame(file.Path))
		for j, q := range file.Questions {
			a := answers[i][j]
			out.WriteString(frame(q.Operation) + frame(fmt.Sprint(q.Selector.Start)) + frame(fmt.Sprint(q.Selector.End)) + frame(q.Selector.Kind.String()) + frame(a.Value) + frame(a.Error))
		}
	}
	return out.String()
}

type batchMeasurement struct {
	N, Files, Owners, Questions, Errors, Round int
	Generation                                 uint64
	LoadNS, RunNS                              int64
	ReplaySHA256                               string
	ReverseOrderVerified                       bool
}

func measureBatches(config, manifest, out string) {
	data, err := os.ReadFile(manifest)
	must(err)
	paths := strings.Split(strings.TrimSpace(string(data)), "\n")
	records := []batchMeasurement{}
	baseline := ""
	for round := 0; round < 3; round++ {
		for _, n := range []int{1, 2, 12} {
			start := time.Now()
			p := batchProgram(config, n)
			load := time.Since(start).Nanoseconds()
			plan := []plannedFile{}
			for _, path := range paths {
				file := p.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path)))
				if file == nil || len(file.Diagnostics()) > 0 {
					panic("invalid fixture file")
				}
				node := probeNode(file)
				q := batchQuestion{Selector: batchSelector{node.Pos(), node.End(), node.Kind}, Operation: "type-name"}
				questions := []batchQuestion{}
				for range 1000 {
					questions = append(questions, q)
					questions = append(questions, batchQuestion{Operation: "id-name", FromQuestion: true, OperandQuestion: len(questions) - 1})
				}
				questions = append(questions, batchQuestion{Operation: "intentional-error"}, q)
				plan = append(plan, plannedFile{path, questions})
			}
			session := newBatchSession(p)
			generation := session.Generation
			start = time.Now()
			answers := executePlan(session, plan, false)
			elapsed := time.Since(start).Nanoseconds()
			replay := replayBatch(plan, answers)
			if replayBatch(plan, executePlan(session, plan, true)) != replay {
				panic("reverse completion changed replay")
			}
			if baseline == "" {
				baseline = replay
			} else if replay != baseline {
				panic("file/question replay changed with N or completion order")
			}
			for _, file := range answers {
				if file[len(file)-2].Error == "" || file[len(file)-1].Error != "" || file[len(file)-1].Value != file[0].Value || file[0].Value == "" {
					panic("one question error aborted its batch")
				}
			}
			owners := 0
			errors := 0
			questions := 0
			for _, file := range answers {
				for _, a := range file {
					questions++
					if int(a.ID.Owner) > owners {
						owners = int(a.ID.Owner)
					}
					if a.Error != "" {
						errors++
					}
				}
			}
			if errors != len(paths) {
				panic("per-question error coverage drift")
			}
			records = append(records, batchMeasurement{N: n, Files: len(paths), Owners: owners, Questions: questions, Errors: errors, Round: round, Generation: generation, LoadNS: load, RunNS: elapsed, ReplaySHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(replay))), ReverseOrderVerified: true})
		}
	}
	encoded, err := json.MarshalIndent(records, "", "  ")
	must(err)
	must(os.WriteFile(out, append(encoded, '\n'), 0644))
}
