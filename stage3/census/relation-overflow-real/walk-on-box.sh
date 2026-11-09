#!/usr/bin/env bash
# Canonical entry-first census on one checker. Optional smoke slice: --limit 200.
# House proof environments are release-owned: source env.sh, never rerun setup.sh over it.
# See cloud/env-owner.sh, cloud/setup.sh, and stage3/meter/twice-daily.sh.
# Unit counts retain canonical generic deferral and checker-body skip policies.
set -euo pipefail
repository=$(cd "$(dirname "$0")/../../.." && pwd)
cd "$repository"
limit=0
if [[ $# != 0 ]]; then
    [[ $# == 2 && $1 == --limit && $2 =~ ^[1-9][0-9]*$ ]] || {
        echo 'usage: bash stage3/census/relation-overflow-real/walk-on-box.sh [--limit N]' >&2
        exit 2
    }
    limit=$2
fi
source_override=${ADAMIC_TYPESCRIPT_SOURCE:-}
if [[ -n ${ADAMIC_TOOLS_ENV:-} ]]; then
    [[ -r $ADAMIC_TOOLS_ENV ]] || { echo "missing tools environment: $ADAMIC_TOOLS_ENV" >&2; exit 1; }
    source "$ADAMIC_TOOLS_ENV"
else
    for environment in "${ADAMIC_TOOLS:+$ADAMIC_TOOLS/env.sh}" /home/ahra/adamic-tools/env.sh \
        "$HOME/adamic-tools/env.sh" /workspace/adamic-tools/env.sh /opt/adamic-tools/env.sh; do
        if [[ -r $environment ]]; then
            echo "walk: sourcing $environment" >&2
            source "$environment"
            break
        fi
    done
fi
[[ -z $source_override ]] || export ADAMIC_TYPESCRIPT_SOURCE=$source_override
missing=()
for tool in go gofmt node python3 git; do
    command -v "$tool" > /dev/null 2>&1 || missing+=("$tool")
done
if (( ${#missing[@]} )); then
    printf 'missing required tools: %s\n' "${missing[*]}" >&2
    exit 1
fi
# Use this proof worktree's checked-in workspace, local replacements, and checksums.
export GOWORK="$repository/go.work"
go version >&2
node_version=$(node --version)
[[ $node_version == v24.* ]] || { echo "missing required Node 24: found $node_version" >&2; exit 1; }
pinned=050880ce59e30b356b686bd3144efe24f875ebc8
[[ -n ${ADAMIC_TYPESCRIPT_SOURCE:-} && -f $ADAMIC_TYPESCRIPT_SOURCE/src/tsc/tsc.ts ]] || {
    echo 'missing pinned TypeScript 6.0.3 source: ADAMIC_TYPESCRIPT_SOURCE must name the box gate input checkout' >&2
    exit 1
}
source_commit=$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)
[[ $source_commit == "$pinned" ]] || {
    echo "wrong TypeScript source: $ADAMIC_TYPESCRIPT_SOURCE is $source_commit; need 6.0.3 at $pinned" >&2
    exit 1
}
[[ -z $(git -C "$ADAMIC_TYPESCRIPT_SOURCE" status --porcelain --untracked-files=no) ]] || {
    echo "dirty pinned TypeScript source: $ADAMIC_TYPESCRIPT_SOURCE" >&2; exit 1;
}
printf 'walk: adamic %s, TypeScript 6.0.3 %s at %s\n' "$(git rev-parse HEAD)" "$pinned" "$ADAMIC_TYPESCRIPT_SOURCE" >&2
mkdir -p "${TMPDIR:-/tmp}"
work=$(mktemp -d "${TMPDIR:-/tmp}/relation-overflow-walk.XXXXXX")
printf 'walk: artifacts %s\n' "$work" >&2
runner=
stop() {
    local code=$1
    trap - TERM INT
    if [[ -n $runner ]]; then
        if [[ $code == 130 ]]; then kill -INT "$runner" 2>/dev/null || true
        else kill -TERM "$runner" 2>/dev/null || true; fi
        wait "$runner" || true
    fi
    exit "$code"
}
trap 'stop 143' TERM
trap 'stop 130' INT
# Each child gets its own process group. Forward cancellation to the whole group.
cat > "$work/child.py" <<'PY'
import json,os,pathlib,signal,subprocess,sys,time
log=pathlib.Path(sys.argv[1]); stream=sys.argv[2]=='stream'; args=sys.argv[3:]
started=time.monotonic()
with log.open('wb') as output:
    child=subprocess.Popen(args,stdout=subprocess.PIPE if stream else output,stderr=subprocess.STDOUT,start_new_session=True)
    def forward(sig,frame):
        try:os.killpg(child.pid,sig)
        except ProcessLookupError:pass
    signal.signal(signal.SIGTERM,forward);signal.signal(signal.SIGINT,forward)
    try:
        if stream:
            for line in child.stdout:
                output.write(line);output.flush()
                sys.stdout.buffer.write(line);sys.stdout.buffer.flush()
        code=child.wait()
    finally:
        try:os.killpg(child.pid,signal.SIGTERM)
        except ProcessLookupError:pass
record={'command':args,'child_exit_code':code,'elapsed_seconds':time.monotonic()-started}
log.with_suffix(log.suffix+'.json').write_text(json.dumps(record,indent=2)+'\n')
print(json.dumps(record),file=sys.stderr,flush=True)
if code and not stream:sys.stderr.buffer.write(log.read_bytes());sys.stderr.buffer.flush()
sys.exit(code if code>=0 else 128-code)
PY
run() {
    local label=$1 mode=$2 code
    shift 2
    printf 'walk: %s\n' "$label" >&2
    python3 "$work/child.py" "$work/$label.log" "$mode" "$@" &
    runner=$!
    set +e
    wait "$runner"
    code=$?
    set -e
    runner=
    return "$code"
}
# Match setup.sh's exact gitlinks. Reuse the house checkout when this is a fresh proof worktree.
if [[ ! -f cohere/TypeScript/tsc/go.mod || ! -f cohere/TypeScript-shim/checker/go.mod ]]; then
    echo 'walk: missing cohere loader submodules; initializing recorded gitlinks' >&2
    common=$(git rev-parse --path-format=absolute --git-common-dir)
    shared=$(dirname "$common")
    cohere_cache= typescript_cache=
    if [[ -d $common/modules/cohere/objects && -d $common/modules/cohere/modules/TypeScript/objects ]]; then
        cohere_cache=$common/modules/cohere
        typescript_cache=$common/modules/cohere/modules/TypeScript
    elif [[ -f $shared/cohere/go.mod && -f $shared/cohere/TypeScript/tsc/go.mod ]]; then
        cohere_cache=$shared/cohere
        typescript_cache=$shared/cohere/TypeScript
    fi
    restored=false
    if [[ -n $cohere_cache ]]; then
        if run submodules-cache file git -c protocol.file.allow=always \
            -c "url.file://$cohere_cache.insteadOf=https://github.com/system-inc/cohere.git" \
            -c "url.file://$typescript_cache.insteadOf=https://github.com/system-inc/TypeScript.git" \
            submodule update --init --recursive --depth 1; then
            restored=true
        else
            echo 'walk: cohere gitlinks unavailable in box cache; retrying recorded origins' >&2
            run submodules-sync file git submodule sync --recursive
        fi
    fi
    if [[ $restored == false ]]; then
        run submodules file git -c submodule.cohere.url=https://github.com/system-inc/cohere.git \
            submodule update --init --recursive --depth 1 --filter=blob:none
    fi
fi
[[ -f cohere/TypeScript/tsc/go.mod && -f cohere/TypeScript-shim/checker/go.mod ]] || {
    echo 'missing cohere loader submodules: cohere/TypeScript/tsc and cohere/TypeScript-shim/checker' >&2; exit 1;
}
python3 - <<'PY'
import subprocess,sys
for parent,name in [('.', 'cohere'),('cohere','TypeScript')]:
    expected=subprocess.check_output(['git','-C',parent,'ls-tree','HEAD',name],text=True).split()[2]
    actual=subprocess.check_output(['git','-C',parent+'/'+name,'rev-parse','HEAD'],text=True).strip()
    if actual!=expected:sys.exit(f'wrong {parent}/{name} submodule: {actual}; required {expected}')
    if subprocess.check_output(['git','-C',parent+'/'+name,'status','--porcelain','--untracked-files=no'],text=True).strip():
        sys.exit(f'dirty {parent}/{name} loader submodule')
PY
# The meter uses the lockfile install, then apply verifies the package byte digest.
# A house proof lays stage3/api/node_modules itself and has no npm; install only when absent.
if [[ -d stage3/api/node_modules ]]; then
    echo 'walk: stage3/api/node_modules laid by the proof; skipping npm ci' >&2
else
    command -v npm > /dev/null 2>&1 || { echo 'missing stage3/api/node_modules and no npm to install it' >&2; exit 1; }
    run packages file npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund
fi
run apply file bash stage3/apply.sh "$work/adapted"
run latent-overlay file python3 stage3/census/latent/make_overlay.py "$repository" "$work/latent"
run entry-overlay file python3 stage3/meter/entry_overlay.py "$work/latent" "$work/entry"

cat > "$work/box_checker.go" <<'GO'
package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Use the original TS2859 rendering, not extra TypeToString queries in an observer.
func (c *Checker) BoxWalkObserve(flag func(string,*Type,*Type,string,string,string), site func(*ast.Node)) {
    c.boxWalkFlag = flag
    c.boxWalkSite = site
}
func (c *Checker) boxWalkCache(relation *Relation) string {
    switch relation {
    case c.subtypeRelation: return "subtypeRelation"
    case c.strictSubtypeRelation: return "strictSubtypeRelation"
    case c.assignableRelation: return "assignableRelation"
    case c.comparableRelation: return "comparableRelation"
    case c.identityRelation: return "identityRelation"
    }
    panic("unrecognized overflow relation")
}
func (c *Checker) boxWalkBegin(source,target *Type,relation *Relation) bool {
    if c.boxWalkFirst || c.boxWalkFlag == nil { return false }
    c.boxWalkFirst = true
    c.boxWalkFlag("first_flag_pending",source,target,c.boxWalkCache(relation),"","")
    return true
}
GO
cat > "$work/box_load.go" <<'GO'
package load

import (
    "reflect"
    "github.com/microsoft/TypeScript/tsc/shim/ast"
    "github.com/microsoft/TypeScript/tsc/shim/checker"
)
var BoxWalkObserver func(string,*checker.Type,*checker.Type,string,string,string)
var BoxWalkSite func(*ast.Node)
func (p *Program) BoxWalkCounts(c *checker.Checker) map[string]any {
    value := reflect.ValueOf(c).Elem()
    caches := map[string]any{}
    for _,name := range append(append([]string{},relationTables...),"enumRelation") {
        field := value.FieldByName(name)
        if !field.IsValid() { panic("missing relation cache: "+name) }
        if field.Kind()==reflect.Pointer {
            if field.IsNil() { caches[name]=map[string]int{"size":0,"flagged_pairs":0};continue }
            field=field.Elem().FieldByName("results")
        }
        if !field.IsValid() || field.Kind()!=reflect.Map {panic("unreadable relation cache: "+name)}
        flagged:=0
        for iter:=field.MapRange();iter.Next(); {if iter.Value().Uint() & (1<<5)!=0 {flagged++}}
        caches[name]=map[string]int{"size":field.Len(),"flagged_pairs":flagged}
    }
    field:=value.FieldByName("diagnostics")
    if !field.IsValid() || field.Type()!=reflect.TypeOf(ast.DiagnosticsCollection{}) {panic("unreadable checker diagnostics")}
    ts2859:=0
    for _,d:=range (*ast.DiagnosticsCollection)(field.Addr().UnsafePointer()).GetDiagnostics() {if d.Code()==2859 {ts2859++}}
    return map[string]any{"caches":caches,"RelationOverflowMarked":RelationOverflowMarked(c),"TS2859_count":ts2859}
}
GO
cat > "$work/box_lower.go" <<'GO'
package lower

import (
    "github.com/microsoft/TypeScript/tsc/shim/ast"
    "github.com/microsoft/TypeScript/tsc/shim/checker"
)
var BoxWalkEnter func(*ast.Node,string)
var BoxWalkDone func(*checker.Checker)
var BoxWalkStop func() bool
GO
cat > "$work/box_main.go" <<'GO'
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "strconv"
    "sync"
    "sync/atomic"
    "time"
    "github.com/microsoft/TypeScript/tsc/shim/ast"
    "github.com/microsoft/TypeScript/tsc/shim/checker"
    "github.com/system-inc/adamic/internal/load"
    "github.com/system-inc/adamic/internal/lower"
)
type position struct { Unit string `json:"unit"`; Phase string `json:"phase"` }
func main() {
    if len(os.Args)!=4 {panic("usage: box-walk entry limit records.jsonl")}
    limit,err:=strconv.ParseUint(os.Args[2],10,64);if err!=nil {panic(err)}
    started:=time.Now()
    var done atomic.Uint64
    var current atomic.Value
    current.Store(position{os.Args[1],"checking"})
    var outputMu sync.Mutex
    say:=func(row map[string]any) {
        row["units_done"]=done.Load();row["elapsed_seconds"]=time.Since(started).Seconds()
        pos:=current.Load().(position);row["unit"]=pos.Unit;row["phase"]=pos.Phase
        data,err:=json.Marshal(row);if err!=nil {panic(err)}
        outputMu.Lock();defer outputMu.Unlock()
        if _,err=fmt.Fprintln(os.Stdout,string(data));err!=nil {panic(err)}
    }
    stop:=make(chan struct{})
    var heartbeat sync.WaitGroup
    heartbeat.Add(1)
    go func(){defer heartbeat.Done();ticker:=time.NewTicker(60*time.Second);defer ticker.Stop()
        for {select {case <-ticker.C:say(map[string]any{"event":"heartbeat"});case <-stop:return}}
    }()
    defer func(){close(stop);heartbeat.Wait()}()
    load.BoxWalkObserver=func(event string,source,target *checker.Type,cache,sourceText,targetText string) {
        row:=map[string]any{"event":event,"cache":cache,"source_id":source.Id(),"target_id":target.Id()}
        if event=="first_flag" {row["source"]=sourceText;row["target"]=targetText}
        say(row)
    }
    var program *load.Program
    var previousScan *ast.Node
    load.BoxWalkSite=func(node *ast.Node) {
        if program==nil || current.Load().(position).Phase!="refusal_scan" {return}
        top:=node
        for top!=nil && top.Parent!=nil && top.Parent.Kind!=ast.KindSourceFile {top=top.Parent}
        if top==nil || top==previousScan || ast.GetSourceFileOfNode(top)==nil {return}
        previousScan=top
        current.Store(position{program.Where(top),"refusal_scan"})
    }
    program,err=load.LatentLoad([]string{os.Args[1]});if err!=nil {panic(err)}
    if len(program.LatentDiagnostics())!=0 {panic(fmt.Sprintf("entry checker rejected: %v",program.LatentDiagnostics()))}
    entry:=program.Files()[0]
    c,release:=program.Checker(context.Background(),entry);release()
    program.LatentEntryReach()
    if program.Files()[0]!=entry {panic("canonical entry order lost")}
    total:=0
    for _,file:=range program.Files() {
        other,release:=program.Checker(context.Background(),file)
        if other!=c {panic("walk uses more than one checker")};release()
        total+=len(file.Statements.Nodes)
    }
    counts:=func(reason string) {
        row:=program.BoxWalkCounts(c);row["event"]="counts";row["reason"]=reason
        row["units_total"]=total;row["limit"]=limit
        if reason=="end" {row["completed"]=int(done.Load())==total;row["limited"]=limit>0 && done.Load()>=limit && int(done.Load())<total}
        say(row)
    }
    say(map[string]any{"event":"start","entry":os.Args[1],"files":len(program.Files()),"units_total":total,"limit":limit,"heartbeat_seconds":60,"checkpoint_every":500})
    counts("checked")
    lower.BoxWalkEnter=func(node *ast.Node,phase string) {
        label:=program.Where(node)+" ["+node.Kind.String()+"]"
        if node.Name()!=nil && ast.IsIdentifier(node.Name()) {label+=" "+node.Name().Text()}
        current.Store(position{label,phase})
    }
    lower.BoxWalkDone=func(other *checker.Checker) {
        if other!=c {panic("unit checker changed")}
        n:=done.Add(1)
        if n%500==0 {counts("interval")}
    }
    lower.BoxWalkStop=func()bool{return limit>0 && done.Load()>=limit}
    stream,err:=os.Create(os.Args[3]);if err!=nil {panic(err)}
    encoder:=json.NewEncoder(stream)
    lower.Latent(context.Background(),program,func(record lower.LatentFile){if err:=encoder.Encode(record);err!=nil {panic(err)}})
    if err:=stream.Close();err!=nil {panic(err)}
    counts("end")
    if limit==0 && int(done.Load())!=total {panic("canonical walk ended before all top-level units")}
    if limit>0 && done.Load()<limit && int(done.Load())!=total {panic("slice ended before requested limit")}
}
GO
cat > "$work/instrument.py" <<'PY'
import json,pathlib,re,sys
repo=pathlib.Path(sys.argv[1]);work=pathlib.Path(sys.argv[2])
replace=json.loads((work/'entry/overlay.json').read_text())['Replace']
def read(name):return pathlib.Path(replace.get(str(repo/name),str(repo/name))).read_text()
def put(name,text):
    target=work/('instrumented_'+name.replace('/','_'));target.write_text(text);replace[str(repo/name)]=str(target)
def once(text,old,new):
    assert text.count(old)==1,(old,text.count(old))
    return text.replace(old,new,1)
checker='cohere/TypeScript/tsc/internal/checker/'
text=read(checker+'checker.go')
text=once(text,'type Checker struct {','type Checker struct {\n boxWalkFirst bool\n boxWalkFlag func(string,*Type,*Type,string,string,string)\n boxWalkSite func(*ast.Node)')
put(checker+'checker.go',text)
text=read(checker+'relater.go')
text=once(text,'\tr := c.getRelater()\n\tr.relation = relation', '''\tif c.boxWalkSite!=nil {node:=errorNode;if node==nil {node=c.currentNode};if node!=nil {c.boxWalkSite(node)}}
\tr := c.getRelater()
\tr.relation = relation''')
mark='c.setRelationResult(relation, id, RelationComparisonResultFailed|RelationComparisonResultComplexityOverflow)'
text=once(text,mark,mark+'\n boxWalkFirst := c.boxWalkBegin(source,target,relation)')
old='c.reportDiagnostic(NewDiagnosticForNode(errorNode, diagnostics.Excessive_complexity_comparing_types_0_and_1, c.TypeToString(source), c.TypeToString(target)), diagnosticOutput)'
text=once(text,old,'''boxWalkSourceText:=c.TypeToString(source)
 boxWalkTargetText:=c.TypeToString(target)
 if boxWalkFirst {c.boxWalkFlag("first_flag",source,target,c.boxWalkCache(relation),boxWalkSourceText,boxWalkTargetText)}
 c.reportDiagnostic(NewDiagnosticForNode(errorNode, diagnostics.Excessive_complexity_comparing_types_0_and_1, boxWalkSourceText, boxWalkTargetText), diagnosticOutput)''')
put(checker+'relater.go',text)
put(checker+'box_walk.go',(work/'box_checker.go').read_text())
text=read('internal/load/load.go')
needle='loaded := &Program{compiler: program, fs: fs}'
text=once(text,needle,needle+'''
 if BoxWalkObserver!=nil {
  c,release:=loaded.Checker(context.Background(),program.GetSourceFiles()[0])
  c.BoxWalkObserve(BoxWalkObserver,BoxWalkSite)
  release()
 }
''')
put('internal/load/load.go',text)
put('internal/load/box_walk.go',(work/'box_load.go').read_text())
text=read('internal/lower/latent_hook.go')
start=text.index('func latentRun(')
text=text[:start]+text[start:].replace('for _, file := range program.Files() {','for _, file := range program.Files() {\n if BoxWalkStop!=nil && BoxWalkStop() {return}\n if BoxWalkEnter!=nil {BoxWalkEnter(file.AsNode(),"refusal_scan")}',1)
start=text.index('for _, node := range file.Statements.Nodes {',start)
body_start=start+len('for _, node := range file.Statements.Nodes {')
end=text.index('\n\t\t}\n\t\trelease()',body_start)
body=text[body_start:end]
body,n=re.subn(r'(?m)^(\s*)continue$',r'\1return',body)
assert n==2,('unit skip sites',n)
new='''
 if BoxWalkStop!=nil && BoxWalkStop() {release();return}
 func(){
  if BoxWalkEnter!=nil {BoxWalkEnter(node,"lowering")}
  defer func(){if BoxWalkDone!=nil {BoxWalkDone(checker)}}()
'''+body+'\n }()'
text=text[:body_start]+new+text[end:]
put('internal/lower/latent_hook.go',text)
put('internal/lower/box_walk.go',(work/'box_lower.go').read_text())
put('stage3/census/latent/tool/main.go',(work/'box_main.go').read_text())
(work/'walk-overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
PY
run instrumentation file python3 "$work/instrument.py" "$repository" "$work"
# Compile sequentially. The only native-checker change is observational instrumentation.
run build file go build -buildvcs=false -overlay="$work/walk-overlay.json" -o "$work/walk" ./stage3/census/latent/tool
# Exclude inherited census mutants and full/speculative modes from a house environment.
while IFS= read -r name; do unset "$name"; done < <(compgen -v LATENT_ || true)
export LATENT_FULL=0 LATENT_SPECULATIVE=0 LATENT_ASSERT_NO_OUTPUT=1
run walk stream "$work/walk" "$work/adapted/src/tsc/tsc.ts" "$limit" "$work/records.jsonl"
