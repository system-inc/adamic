package worker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func kvCapabilityCheck(t *testing.T, name string) error {
	t.Helper()
	p, err := load.Load([]string{filepath.Join("testdata/kv_capability", name+".a")})
	if err != nil {
		t.Fatal(err)
	}
	types, release := p.Checker(context.Background(), p.Files()[0])
	// These are compiler-rule tests, not runtime ingress: the fixture's known
	// environment declaration supplies the checked type directly. The production
	// contract, the bridge's KV check, is tested separately (TestWorkerKVIngressReachesBridgeCheck).
	var envSymbol *ast.Symbol
	for _, n := range p.Files()[0].Statements.Nodes {
		if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
			envSymbol = types.GetSymbolAtLocation(n.Name())
		}
	}
	if envSymbol == nil {
		release()
		t.Fatal("fixture environment declaration missing")
	}
	env := types.GetDeclaredTypeOfSymbol(envSymbol)
	symbol := env.Symbol()
	err = lower.CheckWorkerEnvironment(p, types, env, nil)
	release()
	if err == nil {
		_, ordinary := lower.LowerWorker(context.Background(), p, symbol)
		if kvLoweringStops[name] {
			return ordinary
		}
		if ordinary != nil {
			t.Fatalf("ordinary lowering still stops: %v", ordinary)
		}
		t.Log("ordinary Worker lowering completed")
	}
	return err
}

// kvLoweringStops are the fixtures that stop whichever check stops them: a member use that isn't head
// 3b-1's awaited get, whose lowering error is returned when the environment check lets it by.
var kvLoweringStops = map[string]bool{"get": true, "put": true, "delete": true, "get_stored": true, "get_then": true, "get_optional": true, "get_value": true, "get_conditional": true}

func TestWorkerKVCapabilityStops(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"object_literal", "structural_argument", "parameter_default", "property_initializer", "cast", "let", "module", "return", "union", "optional", "array", "record", "generic", "capture", "callback", "field_write", "parameter_write"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := kvCapabilityCheck(t, name)
			var refused *lower.Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "workers:KvNamespace") {
				t.Fatalf("KV capability boundary %s: want named Refused, got %v", name, err)
			}
			t.Log(err)
		})
	}
}
func TestWorkerKVMembersAndViewsStop(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"get", "put", "delete", "reverse_view"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := kvCapabilityCheck(t, name)
			var stop *lower.NotYet
			if !errors.As(err, &stop) || !strings.Contains(err.Error(), "workers:KvNamespace.") {
				t.Fatalf("KV member stop %s: %v", name, err)
			}
			if name == "reverse_view" && !strings.Contains(stop.What, "a member no host registration names") {
				t.Fatalf("reverse KV view missed host walk: %v", err)
			}
			if name != "reverse_view" && !strings.Contains(stop.What, "workers:KvNamespace."+name+"; a Workers binding's calls wait on head 3b") {
				t.Fatalf("KV member stop %s lost rule 149's words: %v", name, err)
			}
			t.Log(err)
		})
	}
}

// kvLoweredReadings are the KV fixtures that pass the environment check and lower, get_awaited the one
// call among them (head 3b-1).
var kvLoweredReadings = []string{"const_destructure", "parameter_destructure", "local_chain", "optional_read", "alias", "held_awaited", "held_iterated", "get_awaited"}

func TestWorkerKVBindingReadsLower(t *testing.T) {
	t.Parallel()
	for _, name := range kvLoweredReadings {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := kvCapabilityCheck(t, name); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestWorkerKVFieldIdentityIsPreludeSymbol(t *testing.T) {
	t.Parallel()
	p, err := load.Load([]string{"testdata/kv_capability/fake_name.a"})
	if err != nil {
		t.Fatal(err)
	}
	if err = CheckSignature(context.Background(), p); err == nil || !strings.Contains(err.Error(), "must have type exactly string") {
		t.Fatalf("KV field identity: program lookalike admitted: %v", err)
	}

}
func TestWorkerKVImportIsWorkerOnly(t *testing.T) {
	t.Parallel()
	compiler := workerCompilerProduct(t, nil, false)
	entry, err := filepath.Abs("testdata/kv_capability/const_destructure.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"c", "js"} {
		cmd := exec.CommandContext(t.Context(), compiler, mode, entry)
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), "adamic/workers") || !strings.Contains(string(out), "compile it with adamic worker") {
			t.Fatalf("KV %s import: %v: %s", mode, err, out)
		}
	}
}
func TestWorkerKVPurityRefusesOpaqueBindings(t *testing.T) {
	t.Parallel()
	p, err := load.Load([]string{"testdata/kv_capability/pure.a"})
	if err != nil {
		t.Fatal(err)
	}
	types, release := p.Checker(context.Background(), p.Files()[0])
	var symbol *ast.Symbol
	for _, n := range p.Files()[0].Statements.Nodes {
		if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
			symbol = types.GetSymbolAtLocation(n.Name())
		}
	}
	sites := map[ir.SourcePosition]bool{}
	err = lower.CheckWorkerEnvironment(p, types, types.GetDeclaredTypeOfSymbol(symbol), sites)
	release()
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.LowerWorker(context.Background(), p, symbol)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for i, f := range lowered.Functions {
		if f.Name == "observe" {
			found = true
			err = CheckPure(lowered, i, sites)
			if err == nil || !strings.Contains(err.Error(), "impure: holds workers:KvNamespace") {
				t.Fatalf("KV purity: %v", err)
			}
		}
	}
	if !found {
		t.Fatal("purity fixture lost observe")
	}
}

// Ingress is open: a KV field reaches the bridge as a binding, its name in its own list, checked
// there as Workers' KV namespace before the handler runs, and the handler gets the binding itself.
// A KV field read as a string would answer the string check's 500 for a real namespace.
func TestWorkerKVIngressReachesBridgeCheck(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "worker")
	if err := Build("testdata/kv_capability/const_destructure.a", out); err != nil {
		t.Fatalf("KV ingress contract: build: %v", err)
	}
	bridge, err := os.ReadFile(filepath.Join(out, "worker.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bridge), "const environmentFields = [\"GREETING\"];\nconst environmentKVFields = [\"KV\"];\n") {
		t.Fatalf("KV ingress contract: the bridge doesn't list KV as a binding:\n%s", bridge)
	}
	script := `import assert from 'node:assert/strict';const {default:worker}=await import(process.argv[1]);const namespace={get:async()=>null,put:async()=>{},delete:async()=>{}};const logs=[];console.error=x=>logs.push(x);const good=await worker.fetch(new Request('https://worker.test/ok'),{GREETING:'hi',KV:namespace});assert.equal(good.status,200);assert.equal(await good.text(),'hi/ok');assert.deepEqual(logs,[]);const bad=await worker.fetch(new Request('https://worker.test/ok'),{GREETING:'hi'});assert.equal(bad.status,500);assert.equal(await bad.text(),'internal error');assert.deepEqual(logs,['environment binding "KV" is missing or not a KV namespace']);`
	if report, err := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script, filepath.Join(out, "worker.mjs")).CombinedOutput(); err != nil {
		t.Fatalf("KV ingress contract: %v\n%s", err, report)
	}
}

// Lower on #bxaz0p1: rule 3 sees only an identifier, so a static method taken through a
// member and called through a conditional is held today by unbound-method alone. This pin
// holds that stop by its own words, so if it lifts for statics that don't read this, the
// test reds before rule 3 must cover members.
func TestWorkerKVNeighborStops(t *testing.T) {
	t.Parallel()
	for name, words := range map[string]string{
		"member_conditional": "refuses a method read as a value",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p, err := load.Load([]string{filepath.Join("testdata/kv_capability", name+".a")})
			if err != nil {
				t.Fatal(err)
			}
			types, release := p.Checker(context.Background(), p.Files()[0])
			var envSymbol *ast.Symbol
			for _, n := range p.Files()[0].Statements.Nodes {
				if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
					envSymbol = types.GetSymbolAtLocation(n.Name())
				}
			}
			if envSymbol == nil {
				release()
				t.Fatal("fixture environment declaration missing")
			}
			env := types.GetDeclaredTypeOfSymbol(envSymbol)
			err = lower.CheckWorkerEnvironment(p, types, env, nil)
			release()
			if err != nil {
				t.Fatalf("KV neighbor stop %s: the environment check now stops it, so this pin moves to that rule: %v", name, err)
			}
			_, err = lower.LowerWorker(context.Background(), p, env.Symbol())
			if err == nil || !strings.Contains(err.Error(), words) {
				t.Fatalf("KV neighbor stop %s: its neighbor rule moved, so a KV rule must now cover this road: want %q, got %v", name, words, err)
			}
			t.Log(err)
		})
	}
}

// Language and Lower on #bxaz0p1: a call that hands a binding to a Promise, generator or
// iterator result is consumed in the handler's own frame, judged by the call and never by its
// callee. Kept in a module array, voided or stored, it waits by name in the environment check,
// which this test reads alone: an async IIFE or a conditional callee would otherwise stop later
// at a neighbor rule, and a mutant that let them past the check would go unseen.
// held_awaited and held_iterated, its await and for-of, lower (TestWorkerKVBindingReadsLower).
func TestWorkerKVSuspendedResultsWait(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"held_generator", "held_promise", "held_void", "held_iife", "held_conditional"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p, err := load.Load([]string{filepath.Join("testdata/kv_capability", name+".a")})
			if err != nil {
				t.Fatal(err)
			}
			types, release := p.Checker(context.Background(), p.Files()[0])
			var envSymbol *ast.Symbol
			for _, n := range p.Files()[0].Statements.Nodes {
				if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
					envSymbol = types.GetSymbolAtLocation(n.Name())
				}
			}
			if envSymbol == nil {
				release()
				t.Fatal("fixture environment declaration missing")
			}
			err = lower.CheckWorkerEnvironment(p, types, types.GetDeclaredTypeOfSymbol(envSymbol), nil)
			release()
			var stop *lower.NotYet
			if !errors.As(err, &stop) || !strings.Contains(err.Error(), "isn't awaited or iterated by a for-of in the same frame") {
				t.Fatalf("KV suspended result %s: want the frame NotYet from the environment check, got %v", name, err)
			}
			t.Log(err)
		})
	}
}

// Head 3b-1 admits one road, await <binding>.get(key). Every other road to get stays stopped, each by
// its own stop's words: rule 149's NotYet for a stored, chained, optional or detached get, and the
// environment check's own refusal for a conditional receiver, a KV value outside an admitted slot.
// The unawaited bare call is get in TestWorkerKVMembersAndViewsStop.
func TestWorkerKVGetRoadsStayStopped(t *testing.T) {
	t.Parallel()
	const memberStop = "workers:KvNamespace.get; a Workers binding's calls wait on head 3b"
	for _, road := range []struct {
		name, words string
		refused     bool
	}{
		{"get_stored", memberStop, false},
		{"get_then", memberStop, false},
		{"get_optional", memberStop, false},
		{"get_value", memberStop, false},
		{"get_conditional", "workers:KvNamespace in a value outside an admitted KV slot", true},
	} {
		t.Run(road.name, func(t *testing.T) {
			t.Parallel()
			err := kvCapabilityCheck(t, road.name)
			var stop *lower.NotYet
			var refused *lower.Refused
			kind := errors.As(err, &stop)
			if road.refused {
				kind = errors.As(err, &refused)
			}
			if !kind || !strings.Contains(err.Error(), road.words) {
				t.Fatalf("KV get road %s: want refused=%t naming %q, got %v", road.name, road.refused, road.words, err)
			}
			t.Log(err)
		})
	}
}

// kvLowered is a KV fixture's lowered Worker program, after its environment check.
func kvLowered(t *testing.T, name string) *ir.Program {
	t.Helper()
	p, err := load.Load([]string{filepath.Join("testdata/kv_capability", name+".a")})
	if err != nil {
		t.Fatal(err)
	}
	types, release := p.Checker(context.Background(), p.Files()[0])
	var envSymbol *ast.Symbol
	for _, n := range p.Files()[0].Statements.Nodes {
		if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
			envSymbol = types.GetSymbolAtLocation(n.Name())
		}
	}
	if envSymbol == nil {
		release()
		t.Fatal("fixture environment declaration missing")
	}
	err = lower.CheckWorkerEnvironment(p, types, types.GetDeclaredTypeOfSymbol(envSymbol), nil)
	release()
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.LowerWorker(context.Background(), p, envSymbol)
	if err != nil {
		t.Fatal(err)
	}
	return lowered
}

// kvPromiseHostCalls walks a program's IR by reflection, as CheckPure does, and counts each host call
// whose Returns is ir.Promise that an await settles directly, naming every other one. Lowering makes
// the call await's operand; NormalizeAsync then snapshots every operand of an await into a temporary
// declared just before it, so the operand the walk sees is that temporary, declared with the call,
// read once and only by the await in the very next statement, and never assigned.
func kvPromiseHostCalls(program *ir.Program) (awaited int, stray []string) {
	reads, writes := map[int]int{}, map[int]int{}
	var count func(reflect.Value)
	count = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.CanInterface() {
			switch node := value.Interface().(type) {
			case ir.Read:
				reads[node.Local]++
			case ir.Assign:
				writes[node.Local]++
			}
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				count(value.Elem())
			}
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				count(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				count(value.Index(index))
			}
		}
	}
	bodies := [][]ir.Statement{program.Main}
	for _, function := range program.Functions {
		bodies = append(bodies, function.Body)
	}
	for _, body := range bodies {
		count(reflect.ValueOf(body))
	}
	awaits := func(statement ir.Statement, local int) bool {
		var value ir.Expression
		switch next := statement.(type) {
		case ir.Declare:
			value = next.Value
		case ir.Evaluate:
			value = next.Value
		}
		settled, isAwait := value.(ir.Await)
		if !isAwait {
			return false
		}
		read, isRead := settled.Value.(ir.Read)
		return isRead && read.Local == local
	}
	var walk func(reflect.Value)
	walk = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.CanInterface() {
			switch node := value.Interface().(type) {
			case []ir.Statement:
				for index, statement := range node {
					if declared, isDeclare := statement.(ir.Declare); isDeclare {
						if call, isCall := declared.Value.(ir.NodeHostCall); isCall && call.Returns == ir.Promise && index+1 < len(node) && awaits(node[index+1], declared.Local) && reads[declared.Local] == 1 && writes[declared.Local] == 0 {
							awaited++
							walk(reflect.ValueOf(call.Arguments))
							continue
						}
					}
					walk(reflect.ValueOf(statement))
				}
				return
			case ir.Await:
				if call, isCall := node.Value.(ir.NodeHostCall); isCall && call.Returns == ir.Promise {
					awaited++
					walk(reflect.ValueOf(call.Arguments))
					return
				}
			case ir.NodeHostCall:
				if node.Returns == ir.Promise {
					stray = append(stray, node.Module+"."+node.Member)
				}
			}
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				walk(value.Elem())
			}
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				walk(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				walk(value.Index(index))
			}
		}
	}
	for _, body := range bodies {
		walk(reflect.ValueOf(body))
	}
	return awaited, stray
}

// A host call that settles to a Promise is only ever awaited directly, never held, passed or returned:
// a pending one holds its arguments, a Worker binding among them, until it settles (ir.NodeHostCall).
// Every KV fixture that lowers is walked, and get_awaited must hold its one get.
func TestWorkerKVPromiseHostCallsAreAwaited(t *testing.T) {
	t.Parallel()
	for _, name := range kvLoweredReadings {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			awaited, stray := kvPromiseHostCalls(kvLowered(t, name))
			if len(stray) != 0 {
				t.Fatalf("KV promise host calls %s: not the operand of an await: %v", name, stray)
			}
			want := 0
			if name == "get_awaited" {
				want = 1
			}
			if awaited != want {
				t.Fatalf("KV promise host calls %s: %d awaited, want %d", name, awaited, want)
			}
		})
	}
}

// The built Worker runs under Node with a fake env, as Workers calls it: the awaited get's text or
// null reaches the body, a value of any other type is the bridge's 500 with a log naming
// workers:KvNamespace.get and never the value, and a missing or duck-failing binding is the named
// 500 before the handler runs. GREETING's secret never reaches a body or a log.
func TestWorkerKVGetAwaitedUnderNode(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "worker")
	if err := Build("testdata/kv_capability/get_awaited.a", out); err != nil {
		t.Fatalf("KV get build: %v", err)
	}
	const secret = "s3cr3t-value"
	const script = `import assert from 'node:assert/strict';
const {default: worker} = await import(process.argv[1]);
const [mode, secret] = process.argv.slice(2);
const logs = [];
console.error = (...values) => logs.push(values.join(' '));
const keys = [];
const namespace = (settled) => ({get: async (key) => { keys.push(key); return settled; }, put: async () => {}, delete: async () => {}});
const env = {GREETING: secret};
if (mode === 'text') env.KV = namespace('hello');
if (mode === 'null') env.KV = namespace(null);
if (mode === 'number') env.KV = namespace(42);
if (mode === 'empty_binding') env.KV = {};
const response = await worker.fetch(new Request('https://worker.test/greeting'), env);
const body = await response.text();
const log = logs.join('\n');
assert(!body.includes(secret) && !log.includes(secret), 'the environment secret reached the output');
if (mode === 'text' || mode === 'null') {
	assert.equal(response.status, 200);
	assert.equal(body, mode === 'text' ? 'hello' : 'none');
	assert.deepEqual(keys, ['/greeting']);
	assert.deepEqual(logs, []);
} else if (mode === 'number') {
	assert.equal(response.status, 500);
	assert.equal(body, 'internal error');
	assert.deepEqual(logs, ['workers:KvNamespace.get settled to a value of type number, not a string or null']);
	assert(!log.includes('42'), 'the settled value reached the log');
} else {
	assert.equal(response.status, 500);
	assert.equal(body, 'internal error');
	assert.deepEqual(logs, ['environment binding "KV" is missing or not a KV namespace']);
	assert.deepEqual(keys, []);
}
`
	for _, mode := range []string{"text", "null", "number", "missing", "empty_binding"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			report, err := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script, filepath.Join(out, "worker.mjs"), mode, secret).CombinedOutput()
			if err != nil {
				t.Fatalf("KV get under Node %s: %v\n%s", mode, err, report)
			}
			if strings.Contains(string(report), secret) {
				t.Fatalf("KV get under Node %s: the environment secret reached the output:\n%s", mode, report)
			}
		})
	}
}
