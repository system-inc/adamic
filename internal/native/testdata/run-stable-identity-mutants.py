"""Bounded identity mutants, using Go overlays without editing production sources.
Source the cloud toolchain environment, then run this file. Logs stay under /tmp.
Each product's identity check is also held by its own empty-cache mutation.
"""
from pathlib import Path
import json
import os
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
mutants = [
 ("function-ordinal", "names.go", 'stableName("adamic_function", e.sourceReadable(index), e.functionKey(index, map[int]bool{}))', 'fmt.Sprintf("adamic_function_%d", index)', "TestStableUnitsSurviveMiddleDeclaration"),
 ("module", "names.go", 'strconv.Quote(source.Module)', 'strconv.Quote("")', "TestSourceNamesDoNotMove"),
 ("specialization", "names.go", 'strconv.Quote(source.Specialization)', 'strconv.Quote("")', "TestGenericSpecializationsHaveSourceIdentity"),
 ("role", "names.go", 'strconv.Quote(source.Role)', 'strconv.Quote("")', "TestGeneratedRolesDoNotCollide"),
 ("shape-ordinal", "emit_objects.go", 'stableName("adamic_shape", "layout", key)', 'fmt.Sprintf("adamic_shape_%d", len(e.shapes))', "TestStableShapeIdentity"),
 ("literal", "names.go", 'stableName("adamic_string", "literal", value)', 'stableName("adamic_string", "literal", "")', "TestStableLiteralIdentity"),
 ("class-ordinal", "names.go", 'stableName("adamic_class", e.program.Classes[index-1].Name, e.classKey(index, map[int]bool{}))', 'fmt.Sprintf("adamic_class_%d", index)', "TestStableClassIdentity"),
 ("regexp-ordinal", "regexp.go", 'stableName("adamic_regex", "bytecode", fmt.Sprintf("%q/%q", program.Pattern, program.Flags))', 'stableName("adamic_regex", "bytecode", fmt.Sprintf("%q/%q/%d", program.Pattern, program.Flags, index))', "TestRegexpSymbolsDoNotMove"),
 ("temporary-reset", "emit.go", 'e.temporaries = 0', '// mutant: keep the preceding body counter', "TestStableCacheOwnersAndCounters"),
 ("cache-reset", "emit.go", 'e.caches = 0', '// mutant: keep the preceding cache counter', "TestStableCacheOwnersAndCounters"),
 ("cache-owner", "emit_objects.go", 'owner = e.functionName(e.functionIndex)', 'owner = "main"', "TestStableCacheOwnersAndCounters"),
 ("helper-ordinal", "names.go", 'key := function.Name + fmt.Sprint(function.Closure, function.Returns)', 'key := function.Name + fmt.Sprint(function.Closure, function.Returns, index)', "TestMethodAdaptersShareStableDefinition"),
 ("regexp-literal-rewrite", "regexp.go", 'return out.String()', 'return strings.ReplaceAll(source, old, name)', "TestRegexpSymbolsDoNotMove"),
 ("lossy-escaping", "names.go", 'hex.EncodeToString(digest[:16])', 'hex.EncodeToString(digest[:0])', "TestStableIdentifierBoundsAndEscaping"),
 ("schema-ordinal", "names.go", 'name := stableName(prefix, "data", key)', 'name := e.temporary()', "TestStableSchemaIdentity"),
 ("accessor-ordinal", "names.go", 'return "#accessor:" + e.functionName(function)', 'return "#accessor:" + fmt.Sprint(function)', "TestSourceNamesDoNotMove"),
 ("async-product", "../lower/functions.go", 'function.Source = l.functionSource(declaration)', 'function.Source = l.functionSource(declaration); if function.Async { function.Source = ir.SourceIdentity{} }', "TestProduct_StableIdentityAsyncMethods"),
 ("record-product", "../lower/locals.go", 'l.result.Locals[local].Source = l.sourceIdentity(node)', 'l.result.Locals[local].Source = l.sourceIdentity(node); if l.result.Locals[local].Name == "alias" { l.result.Locals[local].Source = ir.SourceIdentity{} }', "TestProduct_StableIdentityEntriesRecordAlias"),

 ("comparator-counter", "names.go", 'name := stableName(prefix, "data", key)', 'name := stableName(prefix, "data", key + fmt.Sprint(e.temporaries))', "TestStableClosureComparatorIdentity"),
 ("wasi-reader", "wasm_test.go", 'adamic_function_[A-Za-z0-9_]*handleRequest_[a-f0-9]{32}', 'adamic_function_[0-9]+_handleRequest', "TestStableRequestHandlerIdentity"),
 ("region-facts", "emit.go", 'e.regionValues = map[string]bool{}', '// mutant: retain previous-body region facts', "TestStableCacheOwnersAndCounters"),

 ("local-ordinal", "names.go", 'key := owner + "/" + local.Name + "/" + strconv.Itoa(occurrence)', 'key := owner + "/" + local.Name + "/" + strconv.Itoa(occurrence) + "/" + strconv.Itoa(i)', "TestSourceNamesDoNotMove"),
 ("global-ordinal", "names.go", 'key = sourceKey(local.Source) + "/" + local.Name', 'key = strconv.Itoa(i)', "TestSourceNamesDoNotMove"),
 ("module-counter", "emit.go", 'if module, ok := starts[index]; ok {\n\t\t\te.resetCounters()', 'if module, ok := starts[index]; ok {', "TestSourceNamesDoNotMove"),

 ("declaration-ancestry", "names.go", 'fmt.Sprintf("%q", source.Declaration)', 'fmt.Sprintf("%q", []string{})', "TestStableDeclarationAncestry"),
 ("async-helper-ordinal", "async_ordinary.go", 'identity := stableName("fn", e.sourceReadable(index), e.functionKey(index, map[int]bool{}))', 'identity := fmt.Sprint(index)', "TestStableUnitsSurviveMiddleDeclaration"),
 ("decoder-counter", "library_json_decode.go", 'fmt.Sprintf("%#v", schema)', 'fmt.Sprintf("%#v/%d", schema, e.temporaries)', "TestStableDecoderAndCallableDescriptors"),
 ("callable-counter", "view_unions_callable_producers.go", 'fmt.Sprintf("%v/%d/%q/%v", parameters, result, name, members)', 'fmt.Sprintf("%v/%d/%q/%v/%d", parameters, result, name, members, e.temporaries)', "TestStableDecoderAndCallableDescriptors"),
]
# These controls hold emitted definitions, independently of translation-unit layout.
mutants += [
 ("predicate-wrapper-ordinal", "../lower/predicates_proof.go", 'fmt.Sprintf("%s_overload_%d_check", implementation.Name().Text(), ordinal)', 'fmt.Sprintf("%s_overload_%d_check_%d", implementation.Name().Text(), ordinal, wrapper)', "TestStablePredicateHelperUnits"),
]
# unit-filename-ordinal moves with the deferred unit-layout change: filenames are
# deliberately ordinal sixteen-function groups in this naming-only batch.
selected = os.environ.get("ADAMIC_IDENTITY_MUTANTS", "")
if selected:
 mutants = [m for m in mutants if m[0] in selected.split(",")]
directory = Path(tempfile.mkdtemp(prefix="stable-identity-mutants-"))
for name, relative, old, new, test in mutants:
 source = (root / "internal/native" / relative).resolve()
 original = source.read_text()
 assert original.count(old) == 1, (name, "stale mutation")
 changed = directory / (name + ".go")
 changed.write_text(original.replace(old, new))
 overlay = directory / (name + ".json")
 overlay.write_text(json.dumps({"Replace": {str(source): str(changed)}}))
 log = directory / (name + ".log")
 env = dict(os.environ, GOMAXPROCS="4", GOCACHEPROG="", ADAMIC_GOCACHE_OFF="1")
 if name.endswith("-product"):
  env["ADAMIC_BUILD_CACHE_DIR"] = str(directory / (name + "-cache"))
 with log.open("w") as output:
  result = subprocess.run(["go", "test", "-p", "1", "-overlay", str(overlay), "./internal/native", "-run", "^"+test+"$", "-count=1", "-v", "-timeout", "55s"], cwd=root, env=env, stdout=output, stderr=subprocess.STDOUT, timeout=58)
 text = log.read_text()
 assert result.returncode == 1 and "--- FAIL: "+test in text and "[build failed]" not in text and "clang failed" not in text, (name, text)
 print(f"{name}: caught by {test}; {log}", flush=True)
print(f"{len(mutants)} identity mutants caught; production source unchanged")
