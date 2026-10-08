import pathlib,json,sys
base=pathlib.Path(__file__).resolve().parent
responsible={'HasJSDoc':'Parser.make / Go finishNode and JSDoc attachment','DisallowInContext':'Parser.rootAssignment / allowInAssignment / forStatement context propagation','AwaitContext':'Parser.methodBody / arrow / awaitContext propagation','PossiblyContainsDeprecatedTag':'Parser.make / JSDoc comment scanning','Ambient':'Statements.declaration / declare and declaration-file context','PossiblyContainsDynamicImport':'Parser.primary / import expression and SourceFile aggregation','YieldContext':'Parser.methodBody / arrow / yieldContext propagation','DecoratorContext':'Parser.decorator','DisallowConditionalTypesContext':'Parser.type / conditional type context','ThisNodeHasError':'Parser.make / Parser.error / scannerErrors error propagation','PossiblyContainsImportMeta':'Parser.primary / meta-property and SourceFile aggregation','InWithStatement':'Statements.statement / with context','diagnostics':'Parser.error / expect / recoverList (see witness kind)','position':'Parser.make / token / scanner.fullStart (see witness kind)','children/count':'Parser.file / Statements.statement / delimitedList (see first tree divergence)','kind/children':'Parser.primary / Statements.statement / recovery dispatch (see first tree divergence)','payload/list/token-flags':'Parser.literal / templateToken / delimitedList (see witness kind)'}
sections=[]
for arg in sys.argv[1:]:
 p=pathlib.Path(arg);data=json.loads(p.read_text());tag=p.stem
 lines=[f'## {tag}',f'Files selected: **{data["files"]}**; both sides completed: **{data["parsed"]}**; byte-identical canonical trees and diagnostics with all flags: **{data["identical"]}**; identical ignoring node flags: **{data["shape_identical"]}**; adapter/runner failures: **{len(data["failures"])}**.','', '| Divergence class | Files affected | Shortest input bytes | Parser function / context |','|---|---:|---:|---|']
 for c,v in data['classes'].items():lines.append(f'| {c} | {v["files"]} | {v["shortest_bytes"]} | {responsible.get(c.removeprefix("flags/"),"Parser.make / context propagation")} |')
 for c,v in data['classes'].items():
  label=c.replace('/','-');folder=f'census/{tag}/examples'; suffix=pathlib.Path(v['path']).suffix
  lines+=['',f'### {c}',f'Shortest corpus input: `{v["path"]}` ({v["shortest_bytes"]} bytes). [Exact input]({folder}/{label}{suffix}); complete preorder trees: [typescript-go]({folder}/{label}.go.tree.gz), [port]({folder}/{label}.port.tree.gz).','', 'First witness (Go against port):','```text','Go: '+str(v['example']['go']),'Port: '+str(v['example']['port']),'```']
  if len(v['input'])<=2000:lines+=['```typescript',v['input'],'```']
 if data['failures']:lines+=['','Failures are listed separately in `census/'+tag+'.json`; they do not count as parser differences.']
 sections.append('\n'.join(lines))
head='''# Parser scout census — step 25 (#pw720wc)

Base: `ad7bd06632f1`, origin/area/stage1-lint; recovery ancestor `88f4a83d`. No parser changes. The comparison runs the actual port TypeScript on Node through the repository runtime hook and the unmodified pinned typescript-go parser through a Go overlay. This census does not claim native Adamic execution parity.

Canonical byte identity means preorder node kind, byte start/end, full node flags, literal flags, ordered children (depth), list metadata, operator, cooked/raw text, semantic fields, and parser diagnostics (code, byte position/length, category, text). Positions use UTF-8 bytes on both sides. JSDoc attachments are represented by flags; full lazy JSDoc trees are outside `ForEachChild` and need a separate docTypes comparison. Port storage represents OptionalChain and block-scoped declaration bits; other node flags have no stored counterpart and print as unset. Nothing is silently masked in the full comparison.

Classes overlap: a file may occur in several rows. Structural comparison explicitly removes only node flags. “Shortest” means the shortest complete selected corpus file exhibiting that class, not a delta-minimized program. The full input and both complete trees are retained beside the first witness. Function names for flags identify missing production/context propagation. Structural attribution is provisional where several parser paths differ; use the retained tree to inspect the exact node.

Adapter: `census/oracle.go.txt` is a copy of the gate oracle with full flags, safe unterminated no-substitution template slicing, and per-file panic capture. The one-character incomplete-template probe parses and prints on both sides: no adapter failure; only ThisNodeHasError differs. Failure records are excluded from completed comparison denominators. A port panic during parsing is currently reported as a runner failure, rather than fabricated as an AST difference.

Reproduction: fetch cohere and its pinned TypeScript checkout, build the Go overlay at the path in `census/overlay.json` (regenerate absolute paths if relocated), then `python3 census/run.py census/tsc.manifest census/tsc.json /tmp/parser-census-oracle`. `fetch.py` reads `public-pins.tsv`, shallow-fetches each exact SHA, executes no dependency installation or repository scripts, and writes `public.manifest`. Run the same driver on that manifest. Paths in manifests reflect this scout's `/workspace` layout.

## Entry points without an inherited list context

`expression()`, `assignment()`, `rootExpression()`, and `rootAssignment()` are public methods and do not establish an enclosing source/block list. The root wrappers only manage expressionDepth and roots. Called on a fresh Parser they can all recover differently from file parsing, for the same reason as tsprinter: recoverList consults the active outer listContexts to decide when to stop. A legal object method shorthand remains supported; malformed surroundings determine whether recovery consumes the method or returns to the enclosing list. Their ordinary calls from file/statements retain an inherited context.

Additional public paths `allowInAssignment()` and `allowInExpression()` wrap assignment/expression and only change disallowIn; they likewise inherit rather than establish a list context. `type()`, `returnType()`, and `docTypes()` can also start with an empty outer list. `docTypes()` scans raw @type substrings, panics on a missing brace, and uses returnType rather than Go's complete JSDoc grammar; malformed types can diverge, but this is a type/JSDoc entry issue rather than an object-method-specific claim.

External production callers inspected: tsprinter `formatExpression` directly calls expression (the known pinned issue); tsprinter files, lint options/settings, lint reparsing, comments, and ESTree pipeline use file(), which establishes source context. Typeaware constructs Parser and delegates file parsing to its linter. JSX expression calls inherit source/list contexts during file parsing. No second external fresh-Parser expression caller was found in stage1. The report refers to the pinned branch; formatter fix `06c6e7cbf` is not applied here.

'''
(base.parent/'CENSUS.md').write_text(head+'\n\n'.join(sections)+'\n')
