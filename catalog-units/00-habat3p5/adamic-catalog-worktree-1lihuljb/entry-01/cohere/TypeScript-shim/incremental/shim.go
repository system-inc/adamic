// Package incremental re-exports typescript-go's incremental build machinery.
//
// Hand-written rather than generated, for the same reason the parser and format shims are:
// internal/types/tools/generate_shims covers the packages a headless linter needed, and incremental was not among
// them because a linter that always checks everything has no use for change detection.
//
// Note this is internal/execute/incremental, which lives inside the CLI driver rather than
// beside the compiler. That placement is the reason it was easy to miss: a search for
// internal/incremental finds nothing.
//
// What this buys, measured on the ahra tree rather than assumed: tsgo --noEmit --incremental
// costs 2.11s cold and 0.85s warm, so the mechanism is worth ~1.26s to the compiler that
// ships it. What it does NOT buy is a warm program. ReadBuildInfoProgram returns a *Program
// whose `program` field is nil (incremental.go:44-62), GetProgram panics on that path
// (program.go:110), and upstream calls compiler.NewProgram unconditionally right after
// reading the build info (tsc.go:304-310). The build info is a change-detection input, not a
// graph you load instead of building.
//
// One thing that looks alarming and is not, because the first version of this comment got it
// backwards and a reader deserves the corrected form rather than the scary one.
//
// The build info our tree produces carries no semanticDiagnosticsPerFile at all, and neither
// does one from a fresh --noEmit --incremental run: verified on two artifacts, both with
// fileNames 9,982 matching program size. That is normal, and it does not mean the cache is
// inert.
//
// The skip does not replay stored diagnostics. It runs on content hashes. On load,
// setSemanticDiagnostics (buildinfotosnapshot.go:152-159) seeds an empty "no diagnostics"
// entry for every file NOT in changedFilesSet, and changedFilesSet is computed by comparing
// each file's hash against oldFileInfo.version (programtosnapshot.go:104-110). Then
// collectSemanticDiagnosticsOfAffectedFiles (program.go:290-303) checks only files missing
// from that map. So an unchanged file is seeded clean and never re-checked, which is where
// the 2.11s to 0.85s comes from.
//
// The practical consequence for a caller: do not treat an empty semanticDiagnosticsPerFile as
// a cold cache, and do not go looking for stored diagnostics to replay. Hand the machinery a
// build info and a freshly built program, and the signatures do the work.
package incremental

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/execute/incremental"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
)

// Program is the incremental wrapper around a compiler program. Its `program` field is nil
// when it came from ReadBuildInfoProgram, so treat it as change detection rather than a graph.
type Program = incremental.Program

// BuildInfo is the parsed .tsbuildinfo. Read IsValidVersion and IsIncremental before trusting
// it: ReadBuildInfoProgram returns nil for either failure, and a nil here means "rebuild
// everything" rather than "something went wrong".
type BuildInfo = incremental.BuildInfo

// BuildInfoReader reads a .tsbuildinfo from disk through a compiler host.
type BuildInfoReader = incremental.BuildInfoReader

// Host is the incremental machinery's view of the filesystem.
type Host = incremental.Host

// NewBuildInfoReader builds a reader over a compiler host.
func NewBuildInfoReader(host compiler.CompilerHost) BuildInfoReader {
	return incremental.NewBuildInfoReader(host)
}

// ReadBuildInfoProgram loads the previous run's change-detection state. Returns nil when no
// build info exists, its version does not match, or it was not written incrementally.
func ReadBuildInfoProgram(
	config *tsoptions.ParsedCommandLine,
	reader BuildInfoReader,
	host compiler.CompilerHost,
) *Program {
	return incremental.ReadBuildInfoProgram(config, reader, host)
}

// CreateHost adapts a compiler host for the incremental machinery.
func CreateHost(compilerHost compiler.CompilerHost) Host {
	return incremental.CreateHost(compilerHost)
}

// NewProgram wraps a built program in the incremental machinery, computing what changed
// against oldProgram. Pass nil for oldProgram on a cold run.
//
// The wrap is where change detection happens: programToSnapshot hashes every file and
// compares against the old snapshot's versions, so the returned Program already knows its
// changed set before anything is checked.
//
// nestedEmitNow is a clock for project-reference builds and nil is correct for a one-shot
// check. testing populates an extra diagnostics-inspection struct and should stay false
// outside upstream's own tests.
func NewProgram(program *compiler.Program, oldProgram *Program, host Host) *Program {
	return incremental.NewProgram(program, oldProgram, host, nil, false)
}

// EmitBuildInfo writes the build info and nothing else.
//
// This exists because cohere never emits. Under NoEmit, Program.Emit takes a branch that
// writes only the build info (program.go:248-254), which is exactly the artifact a warm run
// needs and none of the JavaScript we do not want. Calling Emit directly at the call site
// would work and would also read like cohere had started emitting output, so the intent is
// named here instead.
//
// Returns the emit result so a caller can surface diagnostics from the write itself; a failed
// build-info write must not pass silently, since the next run would then be cold while
// reporting success.
func EmitBuildInfo(ctx context.Context, program *Program) *compiler.EmitResult {
	return program.Emit(ctx, compiler.EmitOptions{})
}
