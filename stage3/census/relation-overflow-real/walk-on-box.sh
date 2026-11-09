#!/usr/bin/env bash
# Canonical entry-first census on one checker. Optional smoke slice: --limit 200.
# --plain omits relation hooks and source/pair overlays for timing the same walk.
# House proof environments are release-owned: source env.sh, never rerun setup.sh over it.
# See cloud/env-owner.sh, cloud/setup.sh, and stage3/meter/twice-daily.sh.
# Unit counts retain canonical generic deferral and checker-body skip policies.
set -euo pipefail
repository=$(cd "$(dirname "$0")/../../.." && pwd)
cd "$repository"
limit=0
plain=0
while (( $# )); do
    case "$1" in
        --limit) [[ $# -ge 2 && $2 =~ ^[1-9][0-9]*$ ]] || { echo 'missing positive --limit N' >&2; exit 2; }; limit=$2; shift 2 ;;
        --plain) plain=1; shift ;;
        *) echo 'usage: bash stage3/census/relation-overflow-real/walk-on-box.sh [--limit N] [--plain]' >&2; exit 2 ;;
    esac
done
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
cat > "$work/box_attribution.go" <<'GO'
package checker

import (
    "fmt"
    "runtime"
    "slices"
    "sort"
    "strings"
)

type boxRelationKey struct { function,caller,relation string; lower bool }
type boxRelationTotal struct { calls uint64; growth int64 }
type boxStackMemo struct { pcs []uintptr; function,caller string; lower bool }
type boxAttributionState struct {
    depth int
    baseline,last,gaps int64
    rows map[boxRelationKey]*boxRelationTotal
    pcs []uintptr
    memo map[uint64][]boxStackMemo
    outerCalls uint64
    nonGrowingChange,observerGrowth int64
    pair boxPairContext
    pairs map[boxPairKey]*BoxPairRow
    texts map[*Type]string
    storage map[*Type]bool
    storageObservations uint64
}
type boxPairContext struct { receiver,source *Type; direction string }
type boxPairKey struct { key boxRelationKey; receiver,source *Type; direction string }
type BoxPairRow struct {
    Key string `json:"key"`
    Relation string `json:"relation"`
    ReceiverID TypeId `json:"receiver_id"`
    SourceID TypeId `json:"source_id"`
    Receiver string `json:"receiver_member_type"`
    Source string `json:"source_type"`
    Direction string `json:"direction"`
    Calls uint64 `json:"growing_calls"`
    Growth int64 `json:"growth"`
}
type BoxStorageRow struct {
    ID TypeId `json:"type_id"`
    Text string `json:"type"`
    Members int `json:"union_member_count"`
}
type BoxStorageReport struct {
    Distinct int `json:"distinct_sources"`
    Observations uint64 `json:"storage_list_observations"`
    Largest []BoxStorageRow `json:"largest20"`
}
type boxRelationToken struct { entered,outer bool; before int64; relation *Relation }
type BoxRelationRow struct {
    Key string `json:"key"`
    Function string `json:"function"`
    Caller string `json:"caller"`
    Relation string `json:"relation"`
    Calls uint64 `json:"growing_calls"`
    Growth int64 `json:"growth"`
}
type BoxRelationAttribution struct {
    Baseline int64 `json:"assignable_before"`
    Current int64 `json:"assignable_after"`
    Change int64 `json:"cache_change"`
    SumKeyGrowth int64 `json:"sum_key_growth"`
    NoLowerGrowth int64 `json:"hooked_without_lower_frame_growth"`
    UnhookedGrowth int64 `json:"unhooked_growth"`
    Remainder int64 `json:"unattributed_remainder"`
    UnattributedFraction float64 `json:"unattributed_fraction"`
    UnattributedLarge bool `json:"unattributed_large"`
    Assessment string `json:"unattributed_assessment"`
    AccountingOK bool `json:"accounting_ok"`
    OuterCalls uint64 `json:"outer_calls"`
    GrowingCalls uint64 `json:"growing_calls"`
    NonGrowingChange int64 `json:"non_growing_call_change"`
    ObserverGrowth int64 `json:"observer_type_text_growth"`
    Pairs []BoxPairRow `json:"typesMayReach_pairs"`
    Rows []BoxRelationRow `json:"rows"`
    NoLowerRows []BoxRelationRow `json:"hooked_without_lower_frame"`
}
func (c *Checker) BoxAttributionStart() {
    size:=int64(c.assignableRelation.size())
    c.boxAttribution=&boxAttributionState{baseline:size,last:size,rows:map[boxRelationKey]*boxRelationTotal{},pcs:make([]uintptr,64),memo:map[uint64][]boxStackMemo{},pairs:map[boxPairKey]*BoxPairRow{},texts:map[*Type]string{},storage:map[*Type]bool{}}
}
// Resolve inline frames correctly, memoizing exact PC sequences, not just a hash.
func (a *boxAttributionState) frameKey(relation string) boxRelationKey {
    for {
        n:=runtime.Callers(2,a.pcs)
        pcs:=a.pcs[:n]
        hash:=uint64(14695981039346656037)
        for _,pc:=range pcs {hash^=uint64(pc);hash*=1099511628211}
        for _,saved:=range a.memo[hash] {
            if slices.Equal(saved.pcs,pcs) {return boxRelationKey{saved.function,saved.caller,relation,saved.lower}}
        }
        function,caller:="",""
        frames:=runtime.CallersFrames(pcs)
        for {
            frame,more:=frames.Next()
            if function!="" {caller=frame.Function;break}
            if strings.HasPrefix(frame.Function,"github.com/system-inc/adamic/internal/lower.") {function=frame.Function}
            if !more {break}
        }
        if (function=="" || caller=="") && n==len(a.pcs) && len(a.pcs)<4096 {
            a.pcs=make([]uintptr,len(a.pcs)*2);continue
        }
        found:=function!="" && caller!=""
        if !found {function="<no complete internal/lower frame pair>";caller=""}
        a.memo[hash]=append(a.memo[hash],boxStackMemo{append([]uintptr(nil),pcs...),function,caller,found})
        return boxRelationKey{function,caller,relation,found}
    }
}
func (c *Checker) boxRelationEnter(relation *Relation) boxRelationToken {
    a:=c.boxAttribution
    if a==nil {return boxRelationToken{}}
    a.depth++
    token:=boxRelationToken{entered:true,outer:a.depth==1}
    if !token.outer {return token}
    token.before=int64(c.assignableRelation.size())
    // Independently measure cache changes between observed calls, including deletions.
    a.gaps+=token.before-a.last
    a.outerCalls++
    token.relation=relation
    return token
}
func (c *Checker) boxRelationExit(token boxRelationToken) {
    if !token.entered {return}
    a:=c.boxAttribution
    if token.outer {
        if a.depth!=1 {panic("relation attribution depth mismatch")}
        after:=int64(c.assignableRelation.size())
        growth:=after-token.before
        if growth>0 {
            // Capture here, while the original lowering caller is still on the stack.
            key:=a.frameKey(c.boxWalkCache(token.relation))
            row:=a.rows[key];if row==nil {row=&boxRelationTotal{};a.rows[key]=row}
            row.calls++;row.growth+=growth
            if pair:=a.pair;pair.receiver!=nil && strings.HasSuffix(key.function,".(*lowering).typesMayReach") {
                pk:=boxPairKey{key,pair.receiver,pair.source,pair.direction}
                pr:=a.pairs[pk]
                if pr==nil {
                    pr=&BoxPairRow{Key:key.function+" <- "+key.caller,Relation:key.relation,
                        ReceiverID:pair.receiver.Id(),SourceID:pair.source.Id(),Direction:pair.direction,
                        Receiver:c.boxText(pair.receiver),Source:c.boxText(pair.source)}
                    a.pairs[pk]=pr
                }
                pr.Calls++;pr.Growth+=growth
            }
        } else {a.nonGrowingChange+=growth}
        a.last=int64(c.assignableRelation.size())
    }
    a.depth--
    if a.depth<0 {panic("negative relation attribution depth")}
}
func boxAttributionFinish(s *BoxRelationAttribution) {
    s.Change=s.Current-s.Baseline
    s.SumKeyGrowth=0;s.NoLowerGrowth=0;s.GrowingCalls=0
    for _,row:=range s.Rows {s.SumKeyGrowth+=row.Growth;s.GrowingCalls+=row.Calls}
    for _,row:=range s.NoLowerRows {s.NoLowerGrowth+=row.Growth;s.GrowingCalls+=row.Calls}
    s.Remainder=s.NoLowerGrowth+s.UnhookedGrowth+s.NonGrowingChange
    s.AccountingOK=s.Change==s.SumKeyGrowth+s.Remainder
    denominator:=s.Change;if denominator<0 {denominator=-denominator};if denominator<1 {denominator=1}
    magnitude:=s.Remainder;if magnitude<0 {magnitude=-magnitude}
    s.UnattributedFraction=float64(magnitude)/float64(denominator)
    s.UnattributedLarge=magnitude>0 && s.UnattributedFraction>=0.01
    s.Assessment="none"
    if magnitude>0 {s.Assessment="small: less than 1% of absolute net cache change"}
    if s.UnattributedLarge {s.Assessment="large: at least 1% of absolute net cache change (denominator at least 1)"}
    order:=func(rows []BoxRelationRow) {sort.Slice(rows,func(i,j int)bool {
        if rows[i].Growth!=rows[j].Growth {return rows[i].Growth>rows[j].Growth}
        if rows[i].Key!=rows[j].Key {return rows[i].Key<rows[j].Key}
        return rows[i].Relation<rows[j].Relation
    })}
    order(s.Rows);order(s.NoLowerRows)
    sort.Slice(s.Pairs,func(i,j int)bool {
        x,y:=s.Pairs[i],s.Pairs[j]
        if x.Growth!=y.Growth {return x.Growth>y.Growth}
        if x.Key!=y.Key {return x.Key<y.Key}
        if x.ReceiverID!=y.ReceiverID {return x.ReceiverID<y.ReceiverID}
        if x.SourceID!=y.SourceID {return x.SourceID<y.SourceID}
        return x.Direction<y.Direction
    })
}
func (c *Checker) BoxAttributionSnapshot() BoxRelationAttribution {
    a:=c.boxAttribution
    if a==nil {return BoxRelationAttribution{}}
    if a.depth!=0 {panic("attribution snapshot requires an idle checker")}
    s:=BoxRelationAttribution{Baseline:a.baseline,Current:int64(c.assignableRelation.size()),Rows:[]BoxRelationRow{},NoLowerRows:[]BoxRelationRow{}}
    s.UnhookedGrowth=a.gaps+(s.Current-a.last)
    s.OuterCalls=a.outerCalls;s.NonGrowingChange=a.nonGrowingChange;s.ObserverGrowth=a.observerGrowth
    s.Pairs=[]BoxPairRow{}
    for _,pair:=range a.pairs {s.Pairs=append(s.Pairs,*pair)}
    for key,total:=range a.rows {
        row:=BoxRelationRow{key.function+" <- "+key.caller,key.function,key.caller,key.relation,total.calls,total.growth}
        if key.lower {s.Rows=append(s.Rows,row)} else {s.NoLowerRows=append(s.NoLowerRows,row)}
    }
    boxAttributionFinish(&s)
    return s
}
func (c *Checker) BoxAttributionDelta(before,after BoxRelationAttribution) BoxRelationAttribution {
    subtract:=func(left,right []BoxRelationRow) []BoxRelationRow {
        old:=map[boxRelationKey]BoxRelationRow{}
        for _,row:=range left {old[boxRelationKey{function:row.Function,caller:row.Caller,relation:row.Relation}]=row}
        result:=[]BoxRelationRow{}
        for _,row:=range right {
            previous:=old[boxRelationKey{function:row.Function,caller:row.Caller,relation:row.Relation}]
            row.Calls-=previous.Calls;row.Growth-=previous.Growth
            if row.Calls!=0 || row.Growth!=0 {result=append(result,row)}
        }
        return result
    }
    s:=BoxRelationAttribution{Baseline:before.Current,Current:after.Current,
        UnhookedGrowth:after.UnhookedGrowth-before.UnhookedGrowth,
        Rows:subtract(before.Rows,after.Rows),NoLowerRows:subtract(before.NoLowerRows,after.NoLowerRows)}
    s.OuterCalls=after.OuterCalls-before.OuterCalls
    s.NonGrowingChange=after.NonGrowingChange-before.NonGrowingChange
    s.ObserverGrowth=after.ObserverGrowth-before.ObserverGrowth
    oldPairs:=map[string]BoxPairRow{}
    pairKey:=func(r BoxPairRow)string{return fmt.Sprintf("%s/%s/%d/%d/%s",r.Key,r.Relation,r.ReceiverID,r.SourceID,r.Direction)}
    for _,r:=range before.Pairs {oldPairs[pairKey(r)]=r}
    s.Pairs=[]BoxPairRow{}
    for _,r:=range after.Pairs {
        old:=oldPairs[pairKey(r)];r.Calls-=old.Calls;r.Growth-=old.Growth
        if r.Calls!=0 || r.Growth!=0 {s.Pairs=append(s.Pairs,r)}
    }
    boxAttributionFinish(&s)
    return s
}
// Cache text only for growing typesMayReach pairs and the storage report's top 20.
// Formatting runs nested under the hook, and any cache changes it causes go to
// the independently measured observer/unattributed bucket, never to a rule.
func (c *Checker) boxText(t *Type) string {
    a:=c.boxAttribution
    if text,ok:=a.texts[t];ok {return text}
    before:=int64(c.assignableRelation.size())
    if a.depth==0 {a.gaps+=before-a.last}
    a.depth++
    defer func(){
        a.depth--
        delta:=int64(c.assignableRelation.size())-before
        a.gaps+=delta;a.observerGrowth+=delta
        if a.depth==0 {a.last=int64(c.assignableRelation.size())}
    }()
    text:=c.TypeToString(t);a.texts[t]=text
    return text
}
func (c *Checker) BoxTypesMayReachAssignable(receiver,source *Type,receiverToSource bool) bool {
    a:=c.boxAttribution
    if a!=nil {
        old:=a.pair;direction:="source to receiver";if receiverToSource {direction="receiver to source"}
        a.pair=boxPairContext{receiver,source,direction}
        defer func(){a.pair=old}()
    }
    if receiverToSource {return c.IsTypeAssignableTo(receiver,source)}
    return c.IsTypeAssignableTo(source,receiver)
}
func (c *Checker) BoxRecordStorage(types []*Type) {
    a:=c.boxAttribution;if a==nil {return}
    a.storageObservations++
    for _,t:=range types {if t!=nil {a.storage[t]=true}}
}
func (c *Checker) BoxStorageSnapshot() BoxStorageReport {
    a:=c.boxAttribution
    report:=BoxStorageReport{Distinct:len(a.storage),Observations:a.storageObservations,Largest:[]BoxStorageRow{}}
    types:=make([]*Type,0,len(a.storage));for t:=range a.storage {types=append(types,t)}
    members:=func(t *Type)int {if t.Flags()&TypeFlagsUnion!=0 {return len(t.Types())};return 1}
    sort.Slice(types,func(i,j int)bool {x,y:=members(types[i]),members(types[j]);if x!=y {return x>y};return types[i].Id()<types[j].Id()})
    if len(types)>20 {types=types[:20]}
    for _,t:=range types {report.Largest=append(report.Largest,BoxStorageRow{t.Id(),c.boxText(t),members(t)})}
    return report
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
    "path/filepath"
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
    if len(os.Args)!=5 {panic("usage: box-walk entry limit records.jsonl plain")}
    plain:=os.Args[4]=="1"
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
    // The shim's unsafe mirrors must still read the original Checker field offsets.
    if checker.Checker_numberType(c)!=c.GetNumberType() || checker.Checker_stringType(c)!=c.GetStringType() || checker.Checker_booleanType(c)!=c.GetBooleanType() || checker.Checker_undefinedType(c)!=c.GetUndefinedType() {panic("instrumentation changed shim Checker field offsets")}
    program.LatentEntryReach()
    if program.Files()[0]!=entry {panic("canonical entry order lost")}
    total:=0
    for _,file:=range program.Files() {
        other,release:=program.Checker(context.Background(),file)
        if other!=c {panic("walk uses more than one checker")};release()
        total+=len(file.Statements.Nodes)
    }
    // Attribution begins after checking. Snapshots ask no relations; storage text
    // formatting is measured separately in observer_type_text_growth.
    if !plain {c.BoxAttributionStart()}
    at1500:=c.BoxAttributionSnapshot()
    window:=at1500
    have1500,haveWindow:=false,false
    attribution:=func(reason string) {
        if plain {return}
        storage:=c.BoxStorageSnapshot()
        snapshot:=c.BoxAttributionSnapshot()
        n:=done.Load()
        if n==1500 && !have1500 {at1500=snapshot;have1500=true}
        if n==2000 && have1500 && !haveWindow {window=c.BoxAttributionDelta(at1500,snapshot);haveWindow=true}
        table:=filepath.Join(filepath.Dir(os.Args[3]),fmt.Sprintf("relation-attribution-%06d-%s.json",n,reason))
        report:=map[string]any{"units_done":n,"reason":reason,"growth_metric":"positive net assignableRelation entries per outermost call; per-key counts count growing calls only", "cumulative":snapshot,"record_storage":storage,
            "window_1500_2000":map[string]any{"available":haveWindow}}
        if haveWindow {report["window_1500_2000"]=map[string]any{"available":true,"from_units":1500,"to_units":2000,"data":window}}
        data,err:=json.MarshalIndent(report,"","  ");if err!=nil {panic(err)}
        // Publish complete files, then stream the full table so a yielded box's
        // retained stdout includes every key and the window, not only a path.
        file,err:=os.CreateTemp(filepath.Dir(table),".attribution-*");if err!=nil {panic(err)}
        if _,err=file.Write(append(data,'\n'));err!=nil {panic(err)}
        if err=file.Sync();err!=nil {panic(err)}
        if err=file.Close();err!=nil {panic(err)}
        if err=os.Rename(file.Name(),table);err!=nil {panic(err)}
        report["event"]="relation_attribution_table";report["table_path"]=table
        say(report)
        top:=snapshot.Rows;if len(top)>20 {top=top[:20]}
        row:=map[string]any{"event":"relation_attribution","reason":reason,"top20":top,"table_path":table,
            "assignable_before":snapshot.Baseline,"assignable_after":snapshot.Current,"cache_change":snapshot.Change,
            "sum_key_growth":snapshot.SumKeyGrowth,"hooked_without_lower_frame_growth":snapshot.NoLowerGrowth,
            "unhooked_growth":snapshot.UnhookedGrowth,"unattributed_remainder":snapshot.Remainder,
            "unattributed_large":snapshot.UnattributedLarge,"unattributed_assessment":snapshot.Assessment,"accounting_ok":snapshot.AccountingOK,
            "outer_calls":snapshot.OuterCalls,"growing_calls":snapshot.GrowingCalls,
            "non_growing_call_change":snapshot.NonGrowingChange,"observer_type_text_growth":snapshot.ObserverGrowth,
            "record_storage":storage,"window_1500_2000":map[string]any{"available":haveWindow}}
        topPairs:=snapshot.Pairs;if len(topPairs)>20 {topPairs=topPairs[:20]}
        row["typesMayReach_top20"]=topPairs
        if haveWindow {
            topWindowPairs:=window.Pairs;if len(topWindowPairs)>20 {topWindowPairs=topWindowPairs[:20]}
            topWindow:=window.Rows;if len(topWindow)>20 {topWindow=topWindow[:20]}
            row["window_1500_2000"]=map[string]any{"available":true,"from_units":1500,"to_units":2000,"top20":topWindow,"typesMayReach_top20":topWindowPairs,
                "cache_change":window.Change,"sum_key_growth":window.SumKeyGrowth,"unhooked_growth":window.UnhookedGrowth,
                "hooked_without_lower_frame_growth":window.NoLowerGrowth,"unattributed_remainder":window.Remainder,
                "unattributed_large":window.UnattributedLarge,"unattributed_assessment":window.Assessment,"accounting_ok":window.AccountingOK}
        }
        say(row)
        if !snapshot.AccountingOK || (haveWindow && !window.AccountingOK) {panic("relation attribution conservation failed")}
    }
    counts:=func(reason string) {
        row:=program.BoxWalkCounts(c);row["event"]="counts";row["reason"]=reason
        row["units_total"]=total;row["limit"]=limit
        if reason=="end" {row["completed"]=int(done.Load())==total;row["limited"]=limit>0 && done.Load()>=limit && int(done.Load())<total}
        say(row)
        attribution(reason)
    }
    say(map[string]any{"event":"start","entry":os.Args[1],"files":len(program.Files()),"units_total":total,"limit":limit,"heartbeat_seconds":60,"checkpoint_every":500,"shim_offsets_checked":true,"plain":plain})
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
repo=pathlib.Path(sys.argv[1]);work=pathlib.Path(sys.argv[2]);plain=sys.argv[3]=='1'
replace=json.loads((work/'entry/overlay.json').read_text())['Replace']
def read(name):return pathlib.Path(replace.get(str(repo/name),str(repo/name))).read_text()
def put(name,text):
    target=work/('instrumented_'+name.replace('/','_'));target.write_text(text);replace[str(repo/name)]=str(target)
def once(text,old,new):
    assert text.count(old)==1,(old,text.count(old))
    return text.replace(old,new,1)
checker='cohere/TypeScript/tsc/internal/checker/'
text=read(checker+'checker.go')
# The shim has fixed-offset extra_Checker readers. Append, never prepend fields.
start=text.index('type Checker struct {')
end=text.index('\n}',start)
text=text[:end]+'''
 boxWalkFirst bool
 boxWalkFlag func(string,*Type,*Type,string,string,string)
 boxWalkSite func(*ast.Node)
 boxAttribution *boxAttributionState
'''+text[end:]
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
# These cover the public IsTypeAssignableTo method, shim linkname identical/
# strict-subtype methods, ordinary subtype/comparable paths, direct diagnostic
# comparisons, and elaboration. All actual Relaters are entered through Ex.
for name in ([] if plain else ['isTypeRelatedTo','checkTypeRelatedTo','checkTypeRelatedToEx','checkTypeRelatedToAndOptionallyElaborate']):
    pattern=r'(func \(c \*Checker\) '+name+r'\([\s\S]*?\) bool \{)'
    text,n=re.subn(pattern,lambda match:match[1]+'\n boxRelationToken := c.boxRelationEnter(relation)\n defer c.boxRelationExit(boxRelationToken)',text)
    assert n==1,('relation entry',name,n)
assert text.count('r := c.getRelater()')==1
put(checker+'relater.go',text)
put(checker+'box_attribution.go',(work/'box_attribution.go').read_text())
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
if not plain:
    text=read('internal/lower/entries_records.go')
    # Observe exactly the list the production method computes, once per fresh
    # lowering that materializes it. Do not recompute or alter its sources.
    text=once(text,'l.entriesRecordTypesKnown = true',
        'l.entriesRecordTypesKnown = true\n defer func(){l.checker.BoxRecordStorage(l.entriesRecordTypes)}()')
    start=text.index('func (l *lowering) typesMayReach(')
    end=text.index('\n}',start)+2
    body=text[start:end]
    body=once(body,'l.checker.IsTypeAssignableTo(source, receiver)',
        'l.checker.BoxTypesMayReachAssignable(receiver, source, false)')
    body=once(body,'l.checker.IsTypeAssignableTo(receiver, source)',
        'l.checker.BoxTypesMayReachAssignable(receiver, source, true)')
    put('internal/lower/entries_records.go',text[:start]+body+text[end:])
put('stage3/census/latent/tool/main.go',(work/'box_main.go').read_text())
(work/'walk-overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
PY
run instrumentation file python3 "$work/instrument.py" "$repository" "$work" "$plain"
# Compile sequentially. The only native-checker change is observational instrumentation.
run build file go build -buildvcs=false -overlay="$work/walk-overlay.json" -o "$work/walk" ./stage3/census/latent/tool
# Exclude inherited census mutants and full/speculative modes from a house environment.
while IFS= read -r name; do unset "$name"; done < <(compgen -v LATENT_ || true)
export LATENT_FULL=0 LATENT_SPECULATIVE=0 LATENT_ASSERT_NO_OUTPUT=1
run walk stream "$work/walk" "$work/adapted/src/tsc/tsc.ts" "$limit" "$work/records.jsonl" "$plain"
