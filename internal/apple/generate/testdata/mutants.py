# Not parallel: edits implementation files and restores each in finally.
from pathlib import Path
import json
import subprocess

mutants = [
    ("declaration-kind-collision-hidden", "internal/apple/generate/model.go",
     "old != nil && old.declaration.Kind != d.Kind {", "old != nil && old.declaration.Kind != d.Kind && false {", "TestDeclarationKindCollision"),
    ("consumed-parameter-borrowed", "internal/apple/generate/emit.go",
     'if has(p, "NSConsumedAttr") || has(p, "CFConsumedAttr") {', 'if (has(p, "NSConsumedAttr") || has(p, "CFConsumedAttr")) && false {', "TestPatterns"),
    ("manual-release-bound", "internal/apple/generate/emit.go",
     'if kind != naming.CFunction && (n.Name == "retain" || n.Name == "release" || n.Name == "autorelease" || n.Name == "dealloc") {',
     'if kind != naming.CFunction && false && (n.Name == "retain" || n.Name == "release" || n.Name == "autorelease" || n.Name == "dealloc") {', "TestPatterns"),
    ("consumed-receiver-borrowed", "internal/apple/generate/emit.go",
     'if has(n, "NSConsumesSelfAttr") && !constructor {', 'if has(n, "NSConsumesSelfAttr") && !constructor && false {', "TestPatterns"),
    ("retained-c-string-ownership-lost", "internal/apple/generate/emit.go",
     'if result.c == "id" && has(n, "NSReturnsRetainedAttr") {', 'if result.c == "id" && has(n, "NSReturnsRetainedAttr") && false {', "TestCompilerLowersGeneratedModules"),
    ("written-struct-canonicalization-dropped", "internal/apple/generate/emit.go",
     'if native.tag == "rectangle" {', 'if native.tag == "rectangle" && false {', "TestWrittenTagTypeSpellings"),
    ("import-name-collision-accepted", "internal/apple/generate/emit.go",
     "if g.importError != nil {", "if g.importError != nil && false {", "TestImportNameCollision"),
    ("optional-protocol-method-promised", "internal/apple/generate/model.go",
     "return optional > required, nil", "return optional > required && false, nil", "TestPatterns"),
    ("implemented-type-unchecked", "internal/apple/generate/protocol.go",
     'body := " return 0; "', 'body := " return 0; "\n\t\t\tif result == "BOOL" {\n\t\t\t\tresult = "long"\n\t\t\t}', "TestHeaderWitness"),
    ("folded-names-shared", "internal/apple/generate/protocol.go",
     "if len(selectors[name]) > 1 {", "if len(selectors[name]) > 1 && false {", "TestPatterns"),
    ("leaf-ignores-descendants", "internal/apple/generate/leaves.go",
     "references = append(references, own[descendant]...)", "_ = descendant", "TestPatterns"),
    ("leaf-ignores-mutators", "internal/apple/generate/leaves.go",
     'reads := method.Instance && cleanType(method.Result.Qual) != "void" && !mutatorName(method.Name) && !strings.HasPrefix(method.Name, "init")',
     'reads := method.Instance && !strings.HasPrefix(method.Name, "init")', "TestPatterns"),
    ("leaf-ignores-initializers", "internal/apple/generate/leaves.go",
     '&& !strings.HasPrefix(method.Name, "init")', '', "TestPatterns"),
    ("deprecation-ignored", "internal/apple/generate/model.go",
     "deprecated = deprecated || g.deprecatedIn(text, child.Begin.File)", "_ = g.deprecatedIn", "TestPatterns"),
    ("handed-blocks-not-followed", "internal/lower/cycles.go",
     "\t\t\tfor _, handed := range f.l.appleHanded {\n\t\t\t\tnodes = append(nodes, cycleNode{proven: handed})\n\t\t\t}\n", "", "TestDelegateRefusals"),
    ("one-parameter-nullability-dropped", "internal/apple/generate/emit.go",
     "natives = append(natives, native)",
     'if n.Name == "takeText:" { native.tag = "string"; native.adamic = "string" }\n\t\tnatives = append(natives, native)', "TestPatterns"),
    ("options-emitted-as-enum", "internal/apple/generate/model.go",
     'tag = "options"', 'tag = "enum"', "TestPatterns"),
    ("unavailable-method-kept", "internal/apple/generate/emit.go",
     'if gone != "" {\n\t\treturn d, naming.Output{}, nil, nativeType{}, gone, nil',
     'if gone != "" && false {\n\t\treturn d, naming.Output{}, nil, nativeType{}, gone, nil', "TestPatterns"),
    ("file-delta-one-declaration-late", "internal/apple/generate/ast.go",
     "return n, nil", "n.Location.File = s.file\n\treturn n, nil", "TestDocumentOrderFileDelta"),
    ("nondeterministic-module-order", "internal/apple/generate/model.go",
     "sort.Strings(keys)", "sort.Strings(append([]string{}, keys...))", "TestCanonicalModuleOrder"),
    ("global-prefix-dropped", "internal/apple/naming/globals.go",
     'return normalize(framework, true) + name', 'return name', "naming:TestGlobalNameBijection"),
    ("whole-ast-buffered", "internal/apple/generate/ast.go",
     's := &astStream{decoder: json.NewDecoder(input)}',
     'data, readError := io.ReadAll(input); if readError != nil {return readError}; s := &astStream{decoder: json.NewDecoder(bytes.NewReader(data))}', "TestStreamVisitsBeforeEOF"),
]
results = []
for name, filename, old, new, test in mutants:
    path = Path(filename)
    original = path.read_bytes()
    source = original.decode()
    if source.count(old) != 1:
        raise RuntimeError((name, "mutation anchor count", source.count(old)))
    try:
        source = source.replace(old, new, 1)
        if name == "whole-ast-buffered":
            source = source.replace('import (', 'import (\n "bytes"', 1)
        path.write_text(source)
        package = "./internal/apple/generate"
        if test.startswith("naming:"):
            package = "./internal/apple/naming"
            test = test.split(":", 1)[1]
        log = Path("/tmp/apple-generator-mutant-" + name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", package, "-count=1", "-run", "^" + test + "$"], stdout=output, stderr=subprocess.STDOUT)
        failures = [line.strip() for line in log.read_text().splitlines() if "--- FAIL:" in line]
        if result.returncode == 0 or not failures:
            raise RuntimeError((name, "survived or failed to build", log.read_text()))
        results.append(dict(mutant=name, test=test, exit=result.returncode, failures=failures, log=str(log)))
        print(name, "caught by", test, "exit", result.returncode, flush=True)
    finally:
        path.write_bytes(original)
Path("/tmp/apple-generator-mutants.json").write_text(json.dumps(results, indent=2) + "\n")
