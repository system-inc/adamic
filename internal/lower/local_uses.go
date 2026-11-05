package lower

import (
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// localRole is what an IR field holding an int says about a local: that it reads one, writes one,
// names one some other way, or isn't a local at all.
type localRole int

const (
	localNot localRole = iota
	localRead
	localWrite
	localNamed
)

// localField is how one IR field holding an int or a list of them is classified, and, for one that
// isn't a local, why.
type localField struct {
	role localRole
	why  string
}

// localFields classifies every int and []int field of the IR's statements and expressions, and of
// the structs they hold (Binding), keyed by type and field. Every field a lowered program can name a
// local through is here as a read, a write or another naming, so a count of a local's uses
// (localUseCount) misses none; TestLocalFieldsClassifyEveryIntField fails when the ir package gains one
// this doesn't list (Lower's must on #qk8rztp S2c-1: ForOf.Local and Binding.Local write a local
// with no Assign).
var localFields = map[string]localField{
	"Read.Local":                      {role: localRead},
	"Assign.Local":                    {role: localWrite},
	"Declare.Local":                   {role: localWrite},
	"ForOf.Local":                     {role: localWrite},
	"Binding.Local":                   {role: localWrite},
	"Try.CatchLocal":                  {role: localWrite},
	"AllocateEnvironment.Cells":       {role: localNamed},
	"Loop.PerIteration":               {role: localNamed},
	"Block.Closes":                    {role: localNamed},
	"ArrayConcat.GraphTypes":          {why: "allocation sites and graph types"},
	"ArrayFill.GraphTypes":            {why: "allocation sites and graph types"},
	"ArrayFrom.GraphTypes":            {why: "allocation sites and graph types"},
	"ArrayLiteral.GraphTypes":         {why: "allocation sites and graph types"},
	"ArrayMap.GraphTypes":             {why: "allocation sites and graph types"},
	"ArraySlice.GraphTypes":           {why: "allocation sites and graph types"},
	"ArraySplice.GraphTypes":          {why: "allocation sites and graph types"},
	"ArrayVisit.GraphTypes":           {why: "allocation sites and graph types"},
	"MapEntries.GraphTypes":           {why: "allocation sites and graph types"},
	"MapKeys.GraphTypes":              {why: "allocation sites and graph types"},
	"MapNew.GraphTypes":               {why: "allocation sites and graph types"},
	"MapValues.GraphTypes":            {why: "allocation sites and graph types"},
	"ObjectLiteral.GraphTypes":        {why: "allocation sites and graph types"},
	"SetNew.GraphTypes":               {why: "allocation sites and graph types"},
	"SetValues.GraphTypes":            {why: "allocation sites and graph types"},
	"ArrayFill.Site":                  {why: "a write site, numbered for the cycle proof"},
	"ArrayPush.Site":                  {why: "a write site, numbered for the cycle proof"},
	"ArraySplice.Site":                {why: "a write site, numbered for the cycle proof"},
	"MapSet.Site":                     {why: "a write site, numbered for the cycle proof"},
	"SetAdd.Site":                     {why: "a write site, numbered for the cycle proof"},
	"SetIndex.Site":                   {why: "a write site, numbered for the cycle proof"},
	"SetProperty.Site":                {why: "a write site, numbered for the cycle proof"},
	"ParallelMap.Site":                {why: "a move site, numbered for its diagnostics"},
	"RecordCall.Site":                 {why: "a record site, numbered for its diagnostics"},
	"RecordCoalesce.Site":             {why: "a record site, numbered for its diagnostics"},
	"RecordLiteral.Site":              {why: "a record site, numbered for its diagnostics"},
	"Await.Site":                      {why: "a suspension site, numbered for its diagnostics"},
	"Yield.Site":                      {why: "a suspension site, numbered for its diagnostics"},
	"IterableOf.Site":                 {why: "a generator argument's site, numbered for its diagnostics"},
	"Concat.Site":                     {why: "a template's or concatenation's site, numbered for its diagnostics"},
	"ArrayFrom.CallbackType":          {why: "the checker's id of a callback's type"},
	"ArrayMap.CallbackType":           {why: "the checker's id of a callback's type"},
	"ArrayReduce.CallbackType":        {why: "the checker's id of a callback's type"},
	"ArraySort.CallbackType":          {why: "the checker's id of a callback's type"},
	"ArrayVisit.CallbackType":         {why: "the checker's id of a callback's type"},
	"MapForEach.CallbackType":         {why: "the checker's id of a callback's type"},
	"CallClosure.FunctionType":        {why: "the checker's id of a function value's type"},
	"Property.ViewTypeID":             {why: "the checker's id of a view's field type"},
	"Property.ViewReceiverTypeID":     {why: "the checker's id of a view's receiver type"},
	"SetProperty.ViewReceiverTypeID":  {why: "the checker's id of a view's receiver type"},
	"ArraySort.Comparator":            {why: "a function's index"},
	"Call.Function":                   {why: "a function's index"},
	"Call.Virtual":                    {why: "a method slot plus one"},
	"CallClosure.Direct":              {why: "a function's index plus one"},
	"GeneratorNext.Function":          {why: "a function's index"},
	"MakeClosure.Function":            {why: "a function's index"},
	"Method.Function":                 {why: "a function's index"},
	"InstanceOf.Class":                {why: "a class's index"},
	"ObjectLiteral.Class":             {why: "a class's index"},
	"Property.Class":                  {why: "a class's index"},
	"SetProperty.Class":               {why: "a class's index"},
	"ArrayJoin.Depth":                 {why: "a nesting depth"},
	"Break.Depth":                     {why: "a nesting depth"},
	"JSONDecodeField.Node":            {why: "an index into a JSON schema's nodes"},
	"JSONDecodeNode.Children":         {why: "indexes into a JSON schema's nodes"},
	"JSONDecodeSchema.Root":           {why: "an index into a JSON schema's nodes"},
	"JSONField.Name":                  {why: "a string constant's index"},
	"JSONField.Slot":                  {why: "an object slot's index"},
	"NodeBufferCall.Encoding":         {why: "an encoding's number"},
	"RegExpNew.Flags":                 {why: "a string constant's index"},
	"RegExpNew.Index":                 {why: "a pattern's index"},
	"RegExpNew.Source":                {why: "a string constant's index"},
	"RegExpReplacement.Kinds":         {why: "replacement part kinds"},
	"RegExpReplacement.RestKind":      {why: "a replacement part kind"},
	"RegExpReplacementGroup.Argument": {why: "a capture group's number"},
	"SourceLine.Line":                 {why: "a source line"},
	"StringConstant.Index":            {why: "a string constant's index"},
}

// localUseCount counts how lowered nodes name a local: its reads, its writes and any other naming,
// by every int field localFields classifies. It looks into the plain structs a node holds (a ForOf's
// Binding) and leaves the statements and expressions a node holds to walk, which visits each on its
// own. A ForOf with a pattern leaves Local unused, a Block's Closes is one more than the local it
// names, and a Try's CatchLocal is -1 when the catch binds nothing.
type localUseCount struct {
	reads, writes, named int
}

func (count *localUseCount) node(node any, local int) {
	value := reflect.ValueOf(node)
	if value.Kind() != reflect.Struct {
		return
	}
	pattern := false
	if loop, isForOf := node.(ir.ForOf); isForOf {
		pattern = len(loop.Pattern) != 0
	}
	count.fields(value, local, pattern)
}

func (count *localUseCount) fields(value reflect.Value, local int, pattern bool) {
	typeName := value.Type().Name()
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		name := value.Type().Field(index).Name
		role := localFields[typeName+"."+name].role
		switch {
		case field.Type() == reflect.TypeOf(0):
			if typeName == "ForOf" && name == "Local" && pattern {
				continue
			}
			named := local
			if typeName == "Block" && name == "Closes" {
				named = local + 1
			}
			if int(field.Int()) == named {
				count.add(role)
			}
		case field.Type() == reflect.TypeOf([]int(nil)):
			for element := 0; element < field.Len(); element++ {
				if int(field.Index(element).Int()) == local {
					count.add(role)
				}
			}
		case field.Kind() == reflect.Struct && !irNode(field.Type()):
			count.fields(field, local, false)
		case field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Struct && !irNode(field.Type().Elem()):
			for element := 0; element < field.Len(); element++ {
				count.fields(field.Index(element), local, false)
			}
		}
	}
}

func (count *localUseCount) add(role localRole) {
	switch role {
	case localRead:
		count.reads++
	case localWrite:
		count.writes++
	case localNamed:
		count.named++
	}
}

// irNode is a statement or an expression, which walk visits on its own.
func irNode(of reflect.Type) bool {
	return of.Implements(expressionType) || of.Implements(statementType)
}
