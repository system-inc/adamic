package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// A step of an optional chain optionalChain guards keeps the parser's ?. and NodeFlagsOptionalChain, though
// in the guard's present arm it runs as a plain step (#cdpxjmb, optional_chain.go). So a function that reads
// either one raw, on a road a guarded chain's steps take, reads the root as optional and the steps as the
// rest of an unguarded chain: an optional read of a held receiver typed as a maybe, or a stop the guard
// had lifted. questionDot and inOptionalChain give the guard's answer. This holds the rule for the next
// reader: every function of this package that reads a node's QuestionDotToken or NodeFlagsOptionalChain is
// in optionalChainReaders with how many reads it makes and why it may make them raw, the way
// TestReferenceWalksResolveShorthandNames holds symbol comparisons. The count holds a listed function to
// the reads its why was read against, so one added to it later fails until it's read too. It reads
// syntax only: a selector naming either field, a method call of QuestionDotToken among them. Mutant
// (scratch-oc): eepPropertyPresence's l.questionDot(node) put back as access.QuestionDotToken != nil, and
// this names it.
func TestOptionalChainReadersGoThroughTheGuard(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	// reads is each function's reads, as file:line in function.
	reads := map[string][]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			name := symbolComparerName(function)
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if selector, isSelector := node.(*ast.SelectorExpr); isSelector && (selector.Sel.Name == "QuestionDotToken" || selector.Sel.Name == "NodeFlagsOptionalChain") {
					position := files.Position(selector.Sel.Pos())
					reads[name] = append(reads[name], position.Filename+":"+strconv.Itoa(position.Line)+" in "+name)
				}
				return true
			})
		}
	}
	var unlisted, recounted []string
	for name, found := range reads {
		listed, isListed := optionalChainReaders[name]
		if !isListed {
			unlisted = append(unlisted, found...)
			continue
		}
		if listed.count != len(found) {
			sort.Strings(found)
			recounted = append(recounted, name+" is listed with "+strconv.Itoa(listed.count)+" and makes "+strconv.Itoa(len(found))+":\n"+strings.Join(found, "\n"))
		}
	}
	sort.Strings(unlisted)
	sort.Strings(recounted)
	if len(unlisted) > 0 {
		t.Errorf("a function reads a node's ?. or NodeFlagsOptionalChain raw, which a guarded optional chain's steps keep though they run as plain steps (#cdpxjmb); use l.questionDot or l.inOptionalChain, or list it in optionalChainReaders with why no guarded chain reaches it with another answer:\n%s", strings.Join(unlisted, "\n"))
	}
	for _, recount := range recounted {
		t.Errorf("optionalChainReaders' count for a function differs from the reads it makes; read each one, and if its why holds them all, list the count it makes, or else use l.questionDot or l.inOptionalChain: %s", recount)
	}
	var stale []string
	for name := range optionalChainReaders {
		if len(reads[name]) == 0 {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	for _, name := range stale {
		t.Errorf("optionalChainReaders names %s, which no longer reads a node's ?. or NodeFlagsOptionalChain; take it out", name)
	}
}

// optionalChainReader is a function that may read a node's ?. or NodeFlagsOptionalChain raw: how many reads it
// makes, and why no guarded chain's step reaches it with an answer other than the guard's.
type optionalChainReader struct {
	count int
	why   string
}

// The whys most readers share.
const (
	// readerGuardHelper is questionDot and inOptionalChain themselves.
	readerGuardHelper = "the guard's own answer: the raw field, unless optionalChain is guarding the node's chain"
	// readerChainShape reads whether a node has a ?. of its own, or the chain's shape, after the guard's answer.
	readerChainShape = "the chain's shape, a ?. the node itself writes, beside inOptionalChain or questionDot, which give the guard's answer"
	// readerBeforeGuard asks about a ?.'s receiver, which optionalChain lowers before its guard opens.
	readerBeforeGuard = "whether node is the receiver of a ?., which optionalChain lowers before its guard opens, where it may be missing as the token says"
	// readerNodeHost is a Node host road, which a guard never reaches.
	readerNodeHost = "a Node host road that refuses or declines ?.; optionalChainEntry turns down a chain with a Node host member, so no guard reaches it"
	// readerDeclines is a road of its own that a written ?. turns down.
	readerDeclines = "declines or refuses its own road where a ?. or a chain is written; in a guard that's the plain road or a stop, never a read of a missing value"
	// readerProof is an analysis over the source that a ?. makes decline.
	readerProof = "a proof over the source as it's written, which a ?. or a chain makes decline; it never lowers a guard's present arm"
	// readerTarget is a write, whose target can't be an optional chain.
	readerTarget = "an assignment's target, which TypeScript never lets be an optional chain"
)

// optionalChainReaders is every function TestOptionalChainReadersGoThroughTheGuard allows to read a node's ?. or
// NodeFlagsOptionalChain raw, by symbolComparerName's name.
var optionalChainReaders = map[string]optionalChainReader{
	"questionDot":                     {1, readerGuardHelper},
	"inOptionalChain":                 {1, readerGuardHelper},
	"property":                        {1, readerChainShape},
	"typedArrayExpression":            {2, readerChainShape},
	"optionalIndexContinuation":       {1, readerChainShape + ": its callers ask inOptionalChain first"},
	"optionalIntrinsicContinues":      {3, readerChainShape + ": whether the chain goes on past a call, which a guarded chain's inner call does"},
	"neverFrameMarker":                {1, readerDeclines + ": a flagged call, guarded or not, cannot carry a frame marker"},
	"directOnlyNeverFunction":         {1, readerDeclines + ": a flagged call, guarded or not, keeps the parameter storage stop"},
	"optionalCall":                    {2, readerChainShape + ": read after inOptionalChain, so only an unguarded chain's call reaches them"},
	"optionalMethodSignature":         {1, readerChainShape + ": a receiver in a chain keeps optionalCall's stop, and optionalChainEntry turns down a root that's a method's optional call"},
	"acceptsUndefined":                {2, readerBeforeGuard},
	"optionalReceiver":                {3, readerBeforeGuard},
	"optionalChainOperand":            {3, readerBeforeGuard},
	"hostMemberValueCallThrough":      {3, readerDeclines + ": a .call through a chain stops where main stops it"},
	"dictionaryViewUse":               {1, readerProof + ": the use walk declines an optional call handed the view, guarded or not, so the answer it keeps by the cast holds whichever guard is active"},
	"hostMemberValueUse":              {1, readerDeclines},
	"hostMemberValueThrough":          {2, readerDeclines},
	"hostMemberDeclared":              {1, readerTarget},
	"hostMemberAssignment":            {1, readerTarget},
	"nodeFSDirectoryCall":             {3, readerNodeHost},
	"nodeFSFile":                      {3, readerNodeHost},
	"nodeStatsFieldReadRefusal":       {1, readerNodeHost},
	"nodeCaptureStackTraceRefusal":    {2, readerNodeHost},
	"nodeGarbageCollectorRefusal":     {2, readerNodeHost},
	"nodePlatformReadRefusal":         {1, readerNodeHost},
	"nodeEnvironmentReadRefusal":      {1, readerNodeHost},
	"nodeRealpathUseRefusal":          {1, readerNodeHost},
	"nodeStdoutAlias":                 {1, readerNodeHost},
	"nodeStdoutWriteRefusal":          {2, readerNodeHost},
	"nodeStdoutIsTTYRefusal":          {1, readerNodeHost},
	"nodeWatchFile":                   {1, readerNodeHost},
	"nodeProcessPath":                 {1, readerNodeHost},
	"nodeStackTraceLimitRefusal":      {1, readerNodeHost},
	"nodeTimer":                       {1, readerNodeHost},
	"nodeWatchCall":                   {1, readerNodeHost},
	"nodeWatchOn":                     {1, readerNodeHost},
	"processPath":                     {2, readerNodeHost},
	"foreignCall":                     {1, readerDeclines + ": an Apple member through ?. stays stopped, inside a guard too"},
	"foreignProperty":                 {1, readerDeclines + ": an Apple member through ?. stays stopped, inside a guard too"},
	"functionBind":                    {4, readerDeclines + ": bind through a chain stays stopped (#cdpxjmb's stays-stopped list)"},
	"bindReceiverRule":                {2, readerDeclines + ": bind through a chain stays stopped"},
	"isArgumentsLength":               {1, readerDeclines},
	"resolveStructuralMethodThunks":   {1, readerDeclines + ": a call through ?. keeps the structural thunks, the safer table"},
	"captureFreeSplit":                {1, readerDeclines},
	"detachedOwnCallReceiver":         {4, readerDeclines},
	"restAdapterCallback":             {1, readerDeclines},
	"functionObservation":             {2, readerDeclines},
	"calledInPlace":                   {1, readerDeclines},
	"generatorCall":                   {1, readerDeclines},
	"iterableParameter":               {1, readerDeclines},
	"handDrivenNext":                  {3, readerDeclines},
	"libraryArrayMethodName":          {2, readerDeclines},
	"libraryArrayObservation":         {1, readerDeclines},
	"libraryArrayObservedMethod":      {1, readerDeclines},
	"isArrayPredicateCall":            {2, readerDeclines},
	"libraryFunctionMember":           {1, readerDeclines},
	"intlCollatorCompare":             {1, readerDeclines},
	"intlCollatorCompareRead":         {1, readerDeclines},
	"libraryMethod":                   {2, readerDeclines + ": a library method read as a value through a chain stays stopped"},
	"stringRawTemplate":               {1, readerDeclines},
	"methodPresence":                  {4, readerDeclines + ": a presence test at the end of a longer chain stays stopped"},
	"hostOrigin":                      {1, readerDeclines},
	"hostOriginContainedUse":          {1, readerDeclines},
	"hostOriginFunction":              {1, readerDeclines},
	"namespaceEnumView":               {2, readerDeclines + ": a namespace is never missing"},
	"namespaceValueMember":            {3, readerDeclines + ": a namespace is never missing"},
	"namespaceValueNamed":             {2, readerDeclines + ": a namespace is never missing"},
	"namespaceExpression":             {2, readerDeclines + ": a namespace is never missing"},
	"recordAccess":                    {2, readerDeclines + ": ?. on a record stays stopped (#15kcz6y)"},
	"recordDefineShape":               {2, readerDeclines},
	"reflectGet":                      {3, readerDeclines},
	"dynamicKeyedGetterExpression":    {1, readerDeclines},
	"regexBuiltin":                    {1, readerDeclines},
	"regexLiteralTestInput":           {2, readerDeclines},
	"enumTagWalk.exactKindCall":       {2, readerProof},
	"enumTagWalk.fieldRead":           {2, readerProof},
	"accessLocalTests":                {2, readerProof},
	"kindSetBody.kindHelperCall":      {2, readerProof},
	"kindSetBody.read":                {2, readerProof},
	"checkedFieldAliasGap":            {1, readerProof},
	"freshValue":                      {1, readerProof},
	"neverRestReference":              {1, readerProof},
	"neverRestUse":                    {1, readerProof},
	"overloadCallbackCall":            {2, readerProof},
	"overloadCallbackProof.target":    {3, readerProof + ": the cached overload verdict follows bodies and reads written calls and callees even while lowering guards their chain; these reads only refuse, so the written form is the conservative answer"},
	"overloadCallbackProof.arrayUse":  {1, readerProof + ": the cached overload verdict follows bodies and reads written array accesses even while lowering guards their chain; these reads only refuse, so the written form is the conservative answer"},
	"overloadParameterCall":           {1, readerProof},
	"overloadMutatorCall":             {2, readerProof},
	"overloadMapShapeOf":              {2, readerProof},
	"predicateFalseFlow.valueCall":    {1, readerProof},
	"predicateProducerProven":         {1, readerProof},
	"predicateFlowProof.checkedValue": {1, readerProof},
	"checkedPredicateInferredPart":    {1, readerProof},
	"predicateUseDirections":          {1, readerProof},
	"censusLengthLeaf":                {1, readerProof},
	"overloadFilterElementAccess":     {1, readerProof},
	"overloadFilterSlice":             {2, readerProof},
	"overloadFilterParameterRead":     {1, readerProof},
	"calleeSourceText":                {3, "spells the callee as the source writes it for a TypeError's message, where V8 keeps a written ?. whatever optionalChain does with the step; it lowers nothing"},

	// Ruling 17's dispatch (#d1chbm7) reads its own tokens raw, so its scan, run before any guard lowers, and its
	// lowering agree: a ?. on the lookup or on x.kind declines the candidate, failing closed, and fn?.(x) is a bare
	// name's call, the whole chain, which no guard runs as a plain step.
	"dispatchCandidate": {2, "the lookup and its key: a ?. on either declines the dispatch, failing closed"},
	"dispatchCalleeUse": {1, "fn?.(x): a call of a bare name, its ?. the whole chain, read as the scan records it"},
	"dispatchCall":      {1, "fn?.(x): the same token the scan read, so the call's no-match arm is the one admitted"},
}
