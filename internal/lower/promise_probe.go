package lower

// A probe for #96pc420, never for landing: what grows promiseIntoVoid's walk past its bound, and what a
// larger bound or a reach prefilter would cost. Each knob is off unless its environment variable is set,
// so with none set the rule runs exactly as on main.
//
//   ADAMIC_PROBE_PROMISE_BOUND=N     the depth bound, 64 when unset
//   ADAMIC_PROBE_PROMISE_REACH=1     answer nil when nothing reachable from source or target can hold a promise
//   ADAMIC_PROBE_PROMISE_CHAIN=path  append each depth-limit stop's site and the pair chain that reached it
//   ADAMIC_PROBE_PROMISE_TIME=path   rewrite the rule's running totals (calls, seconds, pairs) at most once a second

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

var (
	promiseProbeBound = func() int {
		if value, err := strconv.Atoi(os.Getenv("ADAMIC_PROBE_PROMISE_BOUND")); err == nil && value > 0 {
			return value
		}
		return 64
	}()
	promiseProbeReach      = os.Getenv("ADAMIC_PROBE_PROMISE_REACH") == "1"
	promiseProbeChainPath  = os.Getenv("ADAMIC_PROBE_PROMISE_CHAIN")
	promiseProbeTimePath   = os.Getenv("ADAMIC_PROBE_PROMISE_TIME")
	promiseProbeStack      []string
	promiseProbeLimitChain []string
	promiseProbeMaxDepth   int
	promiseProbeReached    = map[*checker.Type]bool{}
	promiseProbeCalls      int
	promiseProbeLimits     int
	promiseProbeReachSkips int
	promiseProbePairs      int
	promiseProbeSeconds    float64
	promiseProbeReachTime  float64
	promiseProbeWritten    time.Time
)

func (l *lowering) promiseProbeName(t *checker.Type) string {
	if t == nil {
		return "<nil>"
	}
	name := l.checker.TypeToString(t)
	if len(name) > 140 {
		name = name[:140] + "..."
	}
	return name
}

// promiseProbePush records a pair on the chain when the chain is asked for, and returns its pop.
func (l *lowering) promiseProbePush(source, target *checker.Type, depth int) func() {
	promiseProbePairs++
	if depth > promiseProbeMaxDepth {
		promiseProbeMaxDepth = depth
	}
	if promiseProbeChainPath == "" {
		return func() {}
	}
	promiseProbeStack = append(promiseProbeStack, l.promiseProbeName(source)+"  =>  "+l.promiseProbeName(target))
	return func() { promiseProbeStack = promiseProbeStack[:len(promiseProbeStack)-1] }
}

func promiseProbeAtLimit() {
	if promiseProbeChainPath != "" && promiseProbeLimitChain == nil {
		promiseProbeLimitChain = append([]string(nil), promiseProbeStack...)
	}
}

// promiseProbeWalk is promiseIntoVoid's entry under the probe: the reach prefilter, the timing and the
// chain written for a depth-limit answer.
func (l *lowering) promiseProbeWalk(node *ast.Node, source, target *checker.Type) *promiseDiscard {
	start := time.Now()
	promiseProbeCalls++
	defer func() {
		promiseProbeSeconds += time.Since(start).Seconds()
		promiseProbeWriteTotals(false)
	}()
	if promiseProbeReach {
		reachStart := time.Now()
		reaches := l.promiseProbeReaches(source) || l.promiseProbeReaches(target)
		promiseProbeReachTime += time.Since(reachStart).Seconds()
		if !reaches {
			promiseProbeReachSkips++
			return nil
		}
	}
	promiseProbeLimitChain = nil
	promiseProbeMaxDepth = 0
	found := l.promiseIntoVoid(source, target, 0, map[[2]*checker.Type]bool{})
	if found != nil && found.depthLimit {
		promiseProbeLimits++
		if promiseProbeChainPath != "" {
			var text strings.Builder
			fmt.Fprintf(&text, "==== %s | bound %d | source %s | target %s\n", l.program.Where(node), promiseProbeBound, l.promiseProbeName(source), l.promiseProbeName(target))
			for index, pair := range promiseProbeLimitChain {
				fmt.Fprintf(&text, "%4d  %s\n", index, pair)
			}
			promiseProbeAppend(promiseProbeChainPath, text.String())
		}
	}
	return found
}

func promiseProbeAppend(path, text string) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(text)
}

func promiseProbeWriteTotals(force bool) {
	if promiseProbeTimePath == "" || (!force && time.Since(promiseProbeWritten) < time.Second) {
		return
	}
	promiseProbeWritten = time.Now()
	text := fmt.Sprintf("calls %d limits %d reachSkips %d pairs %d seconds %.3f reachSeconds %.3f reachMemo %d bound %d reach %v\n",
		promiseProbeCalls, promiseProbeLimits, promiseProbeReachSkips, promiseProbePairs, promiseProbeSeconds, promiseProbeReachTime, len(promiseProbeReached), promiseProbeBound, promiseProbeReach)
	_ = os.WriteFile(promiseProbeTimePath, []byte(text), 0o644)
}

// promiseProbeReaches is whether any type reachable from t through every edge promiseIntoVoid follows,
// either way round, can hold a promise: union and intersection members, a type parameter's constraint,
// type arguments, array elements, call and construct signatures (plain and canonical, returns and every
// parameter) and properties. If neither side reaches one, the walk can't find a discard at any depth. A
// false answer marks everything the query visited false, since each of them reaches only what the root
// reaches; a true answer marks the root alone. Past 200,000 types it answers true.
func (l *lowering) promiseProbeReaches(root *checker.Type) bool {
	if root == nil {
		return false
	}
	if known, ok := promiseProbeReached[root]; ok {
		return known
	}
	visited := map[*checker.Type]bool{}
	stack := []*checker.Type{root}
	found := false
	for len(stack) > 0 && !found {
		t := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if t == nil || visited[t] {
			continue
		}
		if known, ok := promiseProbeReached[t]; ok {
			if known {
				found = true
			}
			continue
		}
		visited[t] = true
		if len(visited) > 200000 || l.mayHoldPromise(t, 0) {
			found = true
			break
		}
		if t.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
			stack = append(stack, t.Types()...)
		}
		if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
			stack = append(stack, l.checker.GetBaseConstraintOfType(t))
		}
		if t.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
			continue
		}
		if t.Flags()&checker.TypeFlagsObject != 0 && t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			stack = append(stack, l.checker.GetTypeArguments(t)...)
		}
		if l.isArrayType(t) || checker.IsTupleType(t) {
			stack = append(stack, l.checker.GetNumberIndexType(t))
		}
		for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
			for _, signature := range l.checker.GetSignaturesOfType(t, kind) {
				for _, each := range []*checker.Signature{signature, checker.Checker_getCanonicalSignature(l.checker, signature)} {
					stack = append(stack, l.checker.GetReturnTypeOfSignature(each))
					for position := range voidRelatedPositions(each, each) {
						stack = append(stack, l.checker.GetTypeAtPosition(each, position))
					}
					// A rest parameter's own type too, so a tuple rest's every element is reached.
					for _, parameter := range each.Parameters() {
						stack = append(stack, l.checker.GetTypeOfSymbol(parameter))
					}
				}
			}
		}
		for _, property := range l.checker.GetPropertiesOfType(t) {
			stack = append(stack, l.checker.GetTypeOfPropertyOfType(t, property.Name))
		}
	}
	if found {
		promiseProbeReached[root] = true
		return true
	}
	for t := range visited {
		promiseProbeReached[t] = false
	}
	return false
}
