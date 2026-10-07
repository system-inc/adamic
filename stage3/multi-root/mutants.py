"""Run isolated Go overlays; never put a mutant in the working tree."""
import json
import pathlib
import subprocess
import sys

repository = pathlib.Path(__file__).resolve().parents[2]
scratch = pathlib.Path(sys.argv[1]).resolve()
scratch.mkdir(parents=True, exist_ok=True)
variants = [
    ('config-roots-as-entries', 'internal/lower/lower.go',
     'for _, file := range program.Entries() {', 'for _, file := range program.Files() {',
     './internal/lower', '^TestProjectEntryAgreesWithNode$'),
    ('unchecked-unimported-config-files', 'internal/load/project.go',
     'program, err := LoadProject(path)', 'program, err := Load([]string{entry})',
     './internal/load', '^TestProjectChecksUnimportedFiles$'),
    ('outside-config-entry', 'internal/load/project.go',
     'return nil, fmt.Errorf("load: entry %s is outside',
     'program.entries = program.files[:1]; return program, nil; return nil, fmt.Errorf("load: entry %s is outside',
     './internal/load', '^TestProjectEntryMustBeConfigured$'),
    ('drop-direct-declarations', 'internal/load/load.go',
     'if projectConfig == nil || !sourceFile.IsDeclarationFile {',
     'if !sourceFile.IsDeclarationFile {', './internal/load', '^TestDirectDeclarationInputStillReportsTypes$'),
    ('declaration-only-guard', 'internal/lower/lower.go', 'if len(files) == 0 {',
     'if false {', './internal/lower', '^TestDeclarationOnlyRootsCannotExecute$'),
    ('missing-adamic-glob-aliases', 'internal/load/source_fs.go',
     'if strings.HasSuffix(name, ".a") && !s.FS.FileExists(path+"/"+name+".ts") {',
     'if false {', './internal/load', '^TestProjectGlobOrderAndAdditionalStrictness$'),
    ('unchecked-javascript', 'internal/load/project.go', 'if options.GetAllowJS() && options.CheckJs != core.TSTrue {',
     'if false {', './internal/load', '^TestProjectCannotDisableCheckingJavaScriptDependencies$'),
    ('wrong-root-order', 'internal/lower/roots.go', 'for _, root := range roots {',
     'for index := len(roots) - 1; index >= 0; index-- { root := roots[index]',
     './internal/lower', '^TestMultipleRootsAgreeWithNode$'),
    ('repeated-shared-modules', 'internal/lower/roots.go', 'if !seen[module] {', 'if true {',
     './internal/lower', '^TestMultipleRootsAgreeWithNode$'),
    ('weakened-required-options', 'internal/load/project.go', 'if *option.value == core.TSFalse {', 'if false {',
     './internal/load', '^TestProjectRefusesWeakenedOptions$'),
    ('enabled-checker-bypass', 'internal/load/project.go', 'if *option.value == core.TSTrue {', 'if false {',
     './internal/load', '^TestProjectRefusesWeakenedOptions$'),
    ('wrong-module', 'internal/load/project.go', 'if options.Module == core.ModuleKindNone {',
     'options.Module = core.ModuleKindESNext; if options.Module == core.ModuleKindNone {',
     './internal/load', '^TestProjectRefusesWeakenedOptions/module$'),
    ('wrong-module-detection', 'internal/load/project.go', 'if options.ModuleDetection == core.ModuleDetectionKindNone {',
     'options.ModuleDetection = core.ModuleDetectionKindForce; if options.ModuleDetection == core.ModuleDetectionKindNone {',
     './internal/load', '^TestProjectRefusesWeakenedOptions/moduleDetection$'),
    ('wrong-module-resolution', 'internal/load/project.go', 'if options.ModuleResolution == core.ModuleResolutionKindUnknown {',
     'options.ModuleResolution = core.ModuleResolutionKindBundler; if options.ModuleResolution == core.ModuleResolutionKindUnknown {',
     './internal/load', '^TestProjectRefusesWeakenedOptions/moduleResolution$'),
    ('old-target', 'internal/load/project.go', 'if options.Target == core.ScriptTargetNone {',
     'options.Target = core.ScriptTargetES2024; if options.Target == core.ScriptTargetNone {',
     './internal/load', '^TestProjectRefusesWeakenedOptions/target$'),
    ('wrong-class-fields', 'internal/load/project.go', 'if options.UseDefineForClassFields == core.TSFalse {',
     'options.UseDefineForClassFields = core.TSTrue; if options.UseDefineForClassFields == core.TSFalse {',
     './internal/load', '^TestProjectRefusesWeakenedOptions/useDefineForClassFields$'),
    ('unsound-project-json', 'internal/load/source_fs.go', 'return prelude[:start] + prelude[end:], true',
     'return strings.ReplaceAll(prelude[:start] + prelude[end:], "space?: unknown): string | undefined;", "space?: unknown): string;"), true',
     './internal/load', '^TestProjectRetainsSoundPreludeWithHostTypes$'),
    ('ambiguous-source-alias', 'internal/load/load.go',
     'if strings.HasSuffix(fileName, ".a.ts") && fs.FS.FileExists(strings.TrimSuffix(fileName, ".ts")) && fs.FS.FileExists(fileName) {',
     'if false {', './internal/load', '^TestProjectRefusesAmbiguousAdamicAliases$'),
    ('missing-fallback-console', 'internal/load/load.go',
     'if projectConfig != nil && !hasHostConsole(program.GetSourceFiles()) {',
     'if false {', './internal/lower', '^TestMultipleRootsAgreeWithNode/project$'),
]
if len(sys.argv) > 2:
    requested = set(sys.argv[2:])
    known = {variant[0] for variant in variants}
    if not requested <= known:
        raise SystemExit(f'unknown mutants: {sorted(requested - known)}')
    variants = [variant for variant in variants if variant[0] in requested]
for name, relative, before, after, package, test in variants:
    original = repository / relative
    source = original.read_text()
    assert source.count(before) == 1, (name, source.count(before))
    replacement = scratch / (name + '.go')
    replacement.write_text(source.replace(before, after))
    overlay = scratch / (name + '.json')
    overlay.write_text(json.dumps({'Replace': {str(original): str(replacement)}}))
    command = ['go', 'test', '-overlay', str(overlay), package, '-run', test, '-count=1', '-timeout=2m']
    with (scratch / (name + '.log')).open('w') as log:
        result = subprocess.run(command, cwd=repository, stdout=log, stderr=subprocess.STDOUT)
    text = (scratch / (name + '.log')).read_text()
    if result.returncode == 0 or '--- FAIL:' not in text or '[build failed]' in text:
        raise SystemExit(f'{name}: not killed by its test; inspect {scratch / (name + ".log")}')
    print(f'{name}: caught (exit {result.returncode})', flush=True)
