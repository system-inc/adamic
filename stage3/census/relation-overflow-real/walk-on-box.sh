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
    "github.com/microsoft/TypeScript/tsc/internal/ast"
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
    origins map[*Type]map[boxSourceOrigin]bool
    storageHistogram bool
    emit func(map[string]any)
    where func(*ast.Node)string
    marked func()bool
    observerDepth int
}
type boxSourceOrigin struct { kind string; node *ast.Node }
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
    c.boxAttribution=&boxAttributionState{baseline:size,last:size,rows:map[boxRelationKey]*boxRelationTotal{},pcs:make([]uintptr,64),memo:map[uint64][]boxStackMemo{},pairs:map[boxPairKey]*BoxPairRow{},texts:map[*Type]string{},storage:map[*Type]bool{},origins:map[*Type]map[boxSourceOrigin]bool{}}
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
    if a.observerDepth>0 {text:=c.TypeToString(t);a.texts[t]=text;return text}
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
    if !a.storageHistogram && a.emit!=nil && len(types)>0 {
        a.storageHistogram=true
        c.boxObserve(func(){c.boxStorageHistogram(types)})
    }
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
func (c *Checker) BoxStorageObserve(emit func(map[string]any),where func(*ast.Node)string,marked func()bool) {
    a:=c.boxAttribution;a.emit=emit;a.where=where;a.marked=marked
}
func (c *Checker) BoxRecordSource(t *Type,node *ast.Node,kind string) *Type {
    a:=c.boxAttribution
    if a!=nil && t!=nil {
        origins:=a.origins[t];if origins==nil {origins=map[boxSourceOrigin]bool{};a.origins[t]=origins}
        origins[boxSourceOrigin{kind,node}]=true
    }
    return t
}
// Separate observation from rule growth, including structural resolution and
// diagnostic comparisons. Nested observer text formatting is charged just once.
func (c *Checker) boxObserve(fn func()) {
    a:=c.boxAttribution
    before:=int64(c.assignableRelation.size())
    idle:=a.depth==0;if idle {a.gaps+=before-a.last}
    a.depth++;a.observerDepth++
    defer func(){
        a.depth--;a.observerDepth--
        delta:=int64(c.assignableRelation.size())-before
        a.gaps+=delta;a.observerGrowth+=delta
        if idle {a.last=int64(c.assignableRelation.size())}
    }()
    fn()
}
func (c *Checker) boxOrigins(t *Type) []map[string]any {
    a:=c.boxAttribution;result:=[]map[string]any{}
    for origin:=range a.origins[t] {result=append(result,map[string]any{"origin":origin.kind,"node":a.where(origin.node)})}
    sort.Slice(result,func(i,j int)bool {x,y:=fmt.Sprint(result[i]["origin"],result[i]["node"]),fmt.Sprint(result[j]["origin"],result[j]["node"]);return x<y})
    return result
}
func boxObjectFlagNames(t *Type) []string {
    names:=[]string{}
    add:=func(mask ObjectFlags,name string){if t.ObjectFlags()&mask!=0 {names=append(names,name)}}
    // Single bits only; aliases reuse bits across object/union/intersection kinds.
    add(ObjectFlagsClass,"Class")
    add(ObjectFlagsInterface,"Interface")
    add(ObjectFlagsReference,"Reference")
    add(ObjectFlagsTuple,"Tuple")
    add(ObjectFlagsAnonymous,"Anonymous")
    add(ObjectFlagsMapped,"Mapped")
    add(ObjectFlagsInstantiated,"Instantiated")
    add(ObjectFlagsObjectLiteral,"ObjectLiteral")
    add(ObjectFlagsEvolvingArray,"EvolvingArray")
    add(ObjectFlagsObjectLiteralPatternWithComputedProperties,"ObjectLiteralPatternWithComputedProperties")
    add(ObjectFlagsReverseMapped,"ReverseMapped")
    add(ObjectFlagsJsxAttributes,"JsxAttributes")
    add(ObjectFlagsJSLiteral,"JSLiteral")
    add(ObjectFlagsFreshLiteral,"FreshLiteral")
    add(ObjectFlagsArrayLiteral,"ArrayLiteral")
    add(ObjectFlagsPrimitiveUnion,"PrimitiveUnion")
    add(ObjectFlagsContainsWideningType,"ContainsWideningType")
    add(ObjectFlagsContainsObjectOrArrayLiteral,"ContainsObjectOrArrayLiteral")
    add(ObjectFlagsNonInferrableType,"NonInferrableType")
    add(ObjectFlagsCouldContainTypeVariablesComputed,"CouldContainTypeVariablesComputed")
    add(ObjectFlagsCouldContainTypeVariables,"CouldContainTypeVariables")
    add(ObjectFlagsMembersResolved,"MembersResolved")
    if t.Flags()&TypeFlagsObject!=0 {add(ObjectFlagsContainsSpread,"ContainsSpread")}
    if t.Flags()&TypeFlagsObject!=0 {add(ObjectFlagsObjectRestType,"ObjectRestType")}
    if t.Flags()&TypeFlagsObject!=0 {add(ObjectFlagsInstantiationExpressionType,"InstantiationExpressionType")}
    if t.Flags()&TypeFlagsObject!=0 {add(ObjectFlagsSingleSignatureType,"SingleSignatureType")}
    if t.Flags()&TypeFlagsObject!=0 {add(ObjectFlagsIsClassInstanceClone,"IsClassInstanceClone")}
    if t.Flags()&TypeFlagsObject!=0 {add(ObjectFlagsIdenticalBaseTypeCalculated,"IdenticalBaseTypeCalculated")}
    if t.Flags()&TypeFlagsObject!=0 {add(ObjectFlagsIdenticalBaseTypeExists,"IdenticalBaseTypeExists")}
    if t.Flags()&TypeFlagsObject!=0 {add(ObjectFlagsFromTypeNode,"FromTypeNode")}
    if t.Flags()&(TypeFlagsUnionOrIntersection|TypeFlagsSubstitution)!=0 {add(ObjectFlagsIsGenericTypeComputed,"IsGenericTypeComputed")}
    if t.Flags()&(TypeFlagsUnionOrIntersection|TypeFlagsSubstitution)!=0 {add(ObjectFlagsIsGenericObjectType,"IsGenericObjectType")}
    if t.Flags()&(TypeFlagsUnionOrIntersection|TypeFlagsSubstitution)!=0 {add(ObjectFlagsIsGenericIndexType,"IsGenericIndexType")}
    if t.Flags()&TypeFlagsUnion!=0 {add(ObjectFlagsContainsIntersections,"ContainsIntersections")}
    if t.Flags()&TypeFlagsUnion!=0 {add(ObjectFlagsIsUnknownLikeUnionComputed,"IsUnknownLikeUnionComputed")}
    if t.Flags()&TypeFlagsUnion!=0 {add(ObjectFlagsIsUnknownLikeUnion,"IsUnknownLikeUnion")}
    if t.Flags()&TypeFlagsUnion!=0 {add(ObjectFlagsIsUniformEnumComputed,"IsUniformEnumComputed")}
    if t.Flags()&TypeFlagsUnion!=0 {add(ObjectFlagsIsUniformEnum,"IsUniformEnum")}
    if t.Flags()&TypeFlagsIntersection!=0 {add(ObjectFlagsIsNeverIntersectionComputed,"IsNeverIntersectionComputed")}
    if t.Flags()&TypeFlagsIntersection!=0 {add(ObjectFlagsIsNeverIntersection,"IsNeverIntersection")}
    if t.Flags()&TypeFlagsIntersection!=0 {add(ObjectFlagsIsConstrainedTypeVariable,"IsConstrainedTypeVariable")}
    if len(names)==0 {names=append(names,"None")};return names
}
func (c *Checker) boxSourceInfo(t *Type,full bool) map[string]any {
    props:=c.GetPropertiesOfType(t);indices:=c.GetIndexInfosOfType(t)
    flags,objects:=t.Flags(),t.ObjectFlags()
    regular:=c.getRegularTypeOfObjectLiteral(t)
    row:=map[string]any{"type_id":t.Id(),"type":c.boxText(t),"origins":c.boxOrigins(t),
        "TypeFlags":FormatTypeFlags(flags),"TypeFlags_bits":uint32(flags),
        "ObjectFlags":boxObjectFlagNames(t),"ObjectFlags_bits":uint32(objects),
        "property_count":len(props),"index_info_count":len(indices),
        "truly_empty":flags&TypeFlagsObject!=0 && len(props)==0 && len(indices)==0,
        "fresh_literal":objects&ObjectFlagsFreshLiteral!=0 && objects&ObjectFlagsObjectLiteral!=0,
        "regular_type_id":regular.Id(),"evolving":objects&ObjectFlagsEvolvingArray!=0,
        "instantiated":objects&ObjectFlagsInstantiated!=0,
        "call_signature_count":len(c.getSignaturesOfType(t,SignatureKindCall)),
        "construct_signature_count":len(c.getSignaturesOfType(t,SignatureKindConstruct))}
    if full {
        properties:=[]map[string]any{}
        for _,p:=range props {
            pt:=c.getTypeOfSymbol(p)
            properties=append(properties,map[string]any{"name":ast.SymbolName(p),"declared_type":c.TypeToStringEx(pt,nil,TypeFormatFlagsNoTruncation,nil),"type_id":pt.Id()})
        }
        sort.Slice(properties,func(i,j int)bool{return properties[i]["name"].(string)<properties[j]["name"].(string)})
        infos:=[]map[string]any{}
        for _,info:=range indices {infos=append(infos,map[string]any{
            "key_type":c.TypeToStringEx(info.KeyType(),nil,TypeFormatFlagsNoTruncation,nil),
            "value_type":c.TypeToStringEx(info.ValueType(),nil,TypeFormatFlagsNoTruncation,nil),"readonly":info.IsReadonly()})}
        row["properties"]=properties;row["index_infos"]=infos
    }
    return row
}
func (c *Checker) boxStorageHistogram(types []*Type) {
    a:=c.boxAttribution
    unique:=map[*Type]bool{};for _,t:=range types {if t!=nil {unique[t]=true}}
    ordered:=make([]*Type,0,len(unique));for t:=range unique {ordered=append(ordered,t)}
    sort.Slice(ordered,func(i,j int)bool{return ordered[i].Id()<ordered[j].Id()})
    groups:=map[string][]map[string]any{};empty,fresh:=0,0
    samples:=[]map[string]any{}
    for _,t:=range ordered {
        row:=c.boxSourceInfo(t,false);text:=row["type"].(string)
        groups[text]=append(groups[text],row)
        if row["truly_empty"].(bool) {empty++};if row["fresh_literal"].(bool) {fresh++}
        if text=="{}" && len(samples)<10 {samples=append(samples,c.boxSourceInfo(t,true))}
    }
    histogram:=[]map[string]any{}
    for text,sources:=range groups {
        // Each source carries its flags/counts/origins: equal printed text need
        // not mean equal checker structure or flags. Never merge those facts.
        histogram=append(histogram,map[string]any{"type":text,"distinct_type_ids":len(sources),"sources":sources})
    }
    sort.Slice(histogram,func(i,j int)bool {
        x,y:=histogram[i]["distinct_type_ids"].(int),histogram[j]["distinct_type_ids"].(int)
        if x!=y {return x>y};return histogram[i]["type"].(string)<histogram[j]["type"].(string)
    })
    a.emit(map[string]any{"event":"record_storage_histogram","collection":"first materialized storage list",
        "distinct_sources":len(unique),"groups":histogram,"truly_empty_count":empty,"fresh_literal_count":fresh,
        "empty_definition":"Object type with zero properties and zero index infos; signature counts also reported",
        "source_origin_labels":[]string{"literal_own_type","contextual_type","variable_declared_type"},"brace_samples":samples})
}
func boxDiagnosticMessages(d *ast.Diagnostic) []map[string]any {
    rows:=[]map[string]any{{"code":d.Code(),"message":d.String()}}
    for _,next:=range d.MessageChain() {rows=append(rows,boxDiagnosticMessages(next)...)}
    return rows
}
func (c *Checker) boxProbe(source,target *Type,node *ast.Node,measure,diagnose bool) map[string]any {
    before:=c.assignableRelation.size()
    markedBefore:=false;if measure {markedBefore=c.boxAttribution.marked()}
    answer:=c.IsTypeAssignableTo(source,target)
    after:=c.assignableRelation.size()
    row:=map[string]any{"assignable":answer,"source_id":source.Id(),"target_id":target.Id()}
    if measure {row["assignable_before"]=before;row["assignable_after"]=after;row["RelationOverflowMarked_before"]=markedBefore;row["RelationOverflowMarked"]=c.boxAttribution.marked()}
    if !diagnose {row["diagnostics_deferred"]=true;return row}
    messages:=[]map[string]any{}
    if !answer {
        var diagnostics []*ast.Diagnostic
        c.checkTypeRelatedToEx(source,target,c.assignableRelation,node,nil,&diagnostics)
        for _,d:=range diagnostics {messages=append(messages,boxDiagnosticMessages(d)...)}
    }
    row["failure_diagnostics"]=messages
    if answer {row["failure_reason"]="none: assignable"} else if len(messages)==0 {row["failure_reason"]="checker returned false without a property/index diagnostic"} else {row["failure_reason"]="see checker diagnostic chain for the failing property/index or non-structural cause"}
    return row
}
func (c *Checker) BoxStorageProbes() {
    a:=c.boxAttribution;if a==nil || a.emit==nil {return}
    // Rank receiver members by growth against brace-printed sources. If a short
    // smoke has none yet, explicitly label a fallback to the top receivers.
    growth:=map[*Type]int64{};bracePairs:=false
    for key,row:=range a.pairs {if row.Source=="{}" {growth[key.receiver]+=row.Growth;bracePairs=true}}
    if !bracePairs {for key,row:=range a.pairs {growth[key.receiver]+=row.Growth}}
    receivers:=make([]*Type,0,len(growth));for t:=range growth {receivers=append(receivers,t)}
    sort.Slice(receivers,func(i,j int)bool {if growth[receivers[i]]!=growth[receivers[j]] {return growth[receivers[i]]>growth[receivers[j]]};return receivers[i].Id()<receivers[j].Id()})
    if len(receivers)>2 {receivers=receivers[:2]}
    if len(receivers)==0 {a.emit(map[string]any{"event":"record_storage_probe","available":false,"reason":"no growing typesMayReach receiver yet"});return}
    selected:=map[*Type]int64{}
    for key,row:=range a.pairs {if key.receiver==receivers[0] && row.Source=="{}" {selected[key.source]+=row.Growth}}
    candidates:=make([]*Type,0,len(selected));for t:=range selected {candidates=append(candidates,t)}
    sort.Slice(candidates,func(i,j int)bool {if selected[candidates[i]]!=selected[candidates[j]] {return selected[candidates[i]]>selected[candidates[j]]};return candidates[i].Id()<candidates[j].Id()})
    // Text only for known storage sources; the histogram already formatted the
    // first list, later lists may introduce new literal types with new IDs.
    c.boxObserve(func(){
        fallback:=[]*Type{}
        for t:=range a.storage {if _,ok:=selected[t];!ok && c.boxText(t)=="{}" {fallback=append(fallback,t)}}
        sort.Slice(fallback,func(i,j int)bool{return fallback[i].Id()<fallback[j].Id()})
        candidates=append(candidates,fallback...);if len(candidates)>10 {candidates=candidates[:10]}
        // Measure the public calls first, before any explanatory diagnostic
        // comparisons can warm them. These are explicit end-of-walk replays.
        measured:=[]map[string]any{}
        if len(candidates)>0 {
            source:=candidates[0];var node *ast.Node;for origin:=range a.origins[source] {node=origin.node;break}
            for _,receiver:=range receivers {measured=append(measured,map[string]any{"receiver_id":receiver.Id(),"receiver":c.boxText(receiver),"source_id":source.Id(),
                "source_to_receiver":c.boxProbe(source,receiver,node,true,false),"receiver_to_source":c.boxProbe(receiver,source,node,true,false)})}
        }
        samples:=[]map[string]any{}
        for _,source:=range candidates {
            info:=c.boxSourceInfo(source,true);var node *ast.Node
            for origin:=range a.origins[source] {node=origin.node;break}
            info["growth_against_receiver"]=selected[source]
            info["source_to_receiver"]=c.boxProbe(source,receivers[0],node,false,true)
            info["receiver_to_source"]=c.boxProbe(receivers[0],source,node,false,true)
            samples=append(samples,info)
        }
        a.emit(map[string]any{"event":"record_storage_probe","available":len(samples)>0,"growing_brace_pairs_observed":bracePairs,
            "receiver_id":receivers[0].Id(),"receiver":c.boxText(receivers[0]),"receiver_growth":growth[receivers[0]],
            "samples":samples,"two_top_receivers_one_source":measured,"observer_queries":true,
            "probe_timing":"end-of-walk replays before explanatory diagnostic comparisons",
            "requested_receivers":2,"receivers_available":len(growth),"receivers_emitted":len(receivers),
            "samples_requested":10,"samples_emitted":len(samples)})
    })
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
    if !plain {
        c.BoxAttributionStart()
        c.BoxStorageObserve(say,program.Where,func()bool{return load.RelationOverflowMarked(c)})
    }
    at1500:=c.BoxAttributionSnapshot()
    window:=at1500
    have1500,haveWindow:=false,false
    attribution:=func(reason string) {
        if plain {return}
        // Probes are explicitly observer work and are reconciled separately.
        if reason=="end" {c.BoxStorageProbes()}
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
        // Publish the full artifact. The bounded JSONL tables below preserve
        // the top 200 pairs per scope when a proof worktree disappears.
        file,err:=os.CreateTemp(filepath.Dir(table),".attribution-*");if err!=nil {panic(err)}
        if _,err=file.Write(append(data,'\n'));err!=nil {panic(err)}
        if err=file.Sync();err!=nil {panic(err)}
        if err=file.Close();err!=nil {panic(err)}
        if err=os.Rename(file.Name(),table);err!=nil {panic(err)}
        // The artifact retains all pairs. Stream bounded JSONL tables, since
        // proof worktrees disappear. Every row is a full typesMayReach pair.
        streamTable:=func(scope string,fromWindow bool) {
            pairs:=snapshot.Pairs;if fromWindow {pairs=window.Pairs}
            total:=len(pairs);if len(pairs)>200 {pairs=pairs[:200]}
            say(map[string]any{"event":"typesMayReach_pair_table","reason":reason,"scope":scope,"pairs_total":total,"pairs_emitted":len(pairs),"limit":200,"table_path":table})
            for i,pair:=range pairs {say(map[string]any{"event":"typesMayReach_pair","reason":reason,"scope":scope,"rank":i+1,"pair":pair})}
        }
        streamTable("cumulative",false)
        if haveWindow {streamTable("window_1500_2000",true)}
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
    text=once(text,'append(l.entriesRecordTypes, l.checker.GetTypeAtLocation(node))',
        'append(l.entriesRecordTypes, l.checker.BoxRecordSource(l.checker.GetTypeAtLocation(node), node, "literal_own_type"))')
    text=once(text,'append(l.entriesRecordTypes, contextual)',
        'append(l.entriesRecordTypes, l.checker.BoxRecordSource(contextual, node, "contextual_type"))')
    text=once(text,'append(l.entriesRecordTypes, l.checker.GetTypeAtLocation(parent.Name()))',
        'append(l.entriesRecordTypes, l.checker.BoxRecordSource(l.checker.GetTypeAtLocation(parent.Name()), parent.Name(), "variable_declared_type"))')
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
