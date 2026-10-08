package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Recognize declarations, including aliased imports, rather than user spellings.
func (l *lowering) nodeFSFile(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	symbol := l.symbol(callee)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return nil, false, nil
	}
	declaration := symbol.Declarations[0]
	memberName := symbol.Name
	source := ast.GetSourceFileOfNode(declaration)
	if callee.Kind == ast.KindPropertyAccessExpression {
		receiver := callee.AsPropertyAccessExpression().Expression
		if l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), "Date") {
			if symbol.Name != "getTime" && symbol.Name != "valueOf" {
				return nil, true, l.notYet(node, memberName+": "+"Date methods other than getTime/valueOf in the fs file host")
			}
			if len(call.Arguments.Nodes) != 0 {
				return nil, true, l.notYet(node, memberName+": "+"Date getTime/valueOf with arguments")
			}
			value, err := l.expression(receiver)
			return ir.NodeFSFile{Operation: "date_time", Arguments: []ir.Expression{value}, Of: ir.Number}, true, err
		}
	}
	if !load.IsNodeLibrary(source) {
		return nil, false, nil
	}
	parent := declaration.Parent
	for parent != nil && parent.Kind != ast.KindModuleDeclaration {
		parent = parent.Parent
	}
	if parent == nil || (parent.Name().Text() != "node:fs" && parent.Name().Text() != "fs") {
		return nil, false, nil
	}
	args := []ir.Expression{}
	operation := ""
	of := ir.Number // Void calls are only admitted as discarded statements.
	switch symbol.Name {
	case "isFile", "isDirectory", "isSymbolicLink", "isBlockDevice", "isCharacterDevice", "isFIFO", "isSocket":
		if declaration.Parent == nil || declaration.Parent.Name() == nil || declaration.Parent.Name().Text() != "StatsBase" {
			return nil, false, nil
		}
		if node.Flags&ast.NodeFlagsOptionalChain != 0 {
			return nil, true, l.notYet(node, "node:fs.StatsBase."+memberName+" through an optional call")
		}
		if callee.Kind != ast.KindPropertyAccessExpression || len(call.Arguments.Nodes) != 0 {
			return nil, true, l.notYet(node, memberName+": "+"a detached Stats method")
		}
		receiver, err := l.expression(callee.AsPropertyAccessExpression().Expression)
		if err != nil {
			return nil, true, err
		}
		// StatsBase and Dirent share methods, including through a union.
		// The runtime receiver chooses its layout, not the declaration order.
		return ir.NodeHostCall{Module: "node:fs", Member: symbol.Name, Arguments: []ir.Expression{receiver}, Returns: ir.Boolean}, true, nil
	case "readFileSync":
		operation, of = "read_file", ir.String
	case "mkdtempSync":
		operation, of = "mkdtemp", ir.String
	case "rmSync":
		operation = "rm"
	case "openSync":
		operation = "open"
	case "readSync":
		if len(call.Arguments.Nodes) != 5 || !l.nodeBufferType(l.checker.GetTypeAtLocation(call.Arguments.Nodes[1]), "Buffer") {
			return nil, true, l.notYet(node, "node:fs.readSync outside the five-argument Buffer overload")
		}
		operation = "read_sync"
	case "writeSync":
		operation = "write"
	case "closeSync":
		operation = "close"
	case "writeFileSync":
		operation = "write_file"
	case "existsSync":
		operation, of = "exists", ir.Boolean
	case "statSync":
		operation, of = "stat", ir.Object
	case "mkdirSync":
		operation, of = "mkdir", ir.String
	case "unlinkSync":
		operation = "unlink"
	case "utimesSync":
		operation = "utimes"
	default:
		return nil, false, nil
	}
	switch operation {
	case "close", "write_file", "unlink", "utimes", "rm":
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		discarded := outer.Parent != nil && outer.Parent.Kind == ast.KindExpressionStatement
		returned := outer.Parent != nil && outer.Parent.Kind == ast.KindReturnStatement && l.function != nil && l.function.Returns == 0
		if !discarded && !returned {
			return nil, true, l.notYet(node, memberName+": "+"fs void calls used as values")
		}
	}
	for index, argument := range call.Arguments.Nodes {
		lowerArgument := l.expression
		if operation == "write" && index == 2 && l.checker.GetTypeAtLocation(argument).Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			// This intrinsic consumes an exact empty position as its existing sentinel; it stores no nullable view.
			lowerArgument = l.value
		}
		value, err := lowerArgument(argument)
		if err != nil {
			if missing, ok := err.(*NotYet); ok {
				err = l.notYet(node, "node:fs."+memberName+": "+missing.What)
			}
			return nil, true, err
		}
		args = append(args, value)
	}
	// Lower only the proven shapes implemented by this host. Broader declared
	// Node overloads produce NotYet rather than reaching C with a wrong type.
	constant := func(s string) ir.Expression { return ir.StringConstant{Index: l.constant(s)} }
	number := func(n float64) ir.Expression { return ir.NumberConstant{Value: n} }
	option := func(value ir.Expression, name string, fallback ir.Expression) (ir.Expression, error) {
		literal, ok := value.(ir.ObjectLiteral)
		if !ok {
			read, reading := value.(ir.Read)
			if reading {
				for _, argument := range call.Arguments.Nodes {
					if !ast.IsIdentifier(ast.SkipParentheses(argument)) || !l.nodeHostExactOptions(argument, node) {
						continue
					}
					symbol := l.symbol(ast.SkipParentheses(argument))
					if local, found := l.locals[symbol]; !found || local != read.Local {
						continue
					}
					field := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(argument), name)
					if field == nil {
						return fallback, nil
					}
					of, known := l.representation(l.checker.GetTypeOfSymbol(field))
					if !known || of != fallback.Type() || field.Flags&ast.SymbolFlagsOptional != 0 || name == "bigint" || name == "encoding" {
						return nil, l.notYet(node, memberName+": "+"fs options with optional fields, dynamic encoding or bigint")
					}
					return ir.Property{Object: value, Name: name, Of: of}, nil
				}
			}
			return nil, l.notYet(node, memberName+": "+"fs options other than a fixed literal or its plain const binding")
		}
		if literal.Spread != nil {
			return nil, l.notYet(node, memberName+": "+"fs options containing a spread")
		}
		// Splitting a literal into runtime parameters must neither drop effects
		// from unused fields nor reorder the effects of its fields.
		for _, field := range literal.Fields {
			switch field.Value.(type) {
			case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Undefined, ir.Null:
			default:
				return nil, l.notYet(node, memberName+": "+"fs option literals containing evaluated expressions; bind a plain options object first")
			}
		}
		for _, field := range literal.Fields {
			if field.Name == name {
				return field.Value, nil
			}
		}
		return fallback, nil
	}
	encoding := func(value ir.Expression) error {
		if _, ok := value.(ir.Undefined); ok {
			return nil
		}
		if _, ok := value.(ir.Null); ok {
			return nil
		}
		if text, ok := value.(ir.StringConstant); ok && (l.result.Strings[text.Index] == "utf8" || l.result.Strings[text.Index] == "utf-8") {
			return nil
		}
		return l.notYet(node, memberName+": "+"fs encodings other than constant utf8/utf-8")
	}
	var err error
	switch operation {
	case "mkdtemp":
		if len(args) > 1 {
			enc := args[1]
			if enc.Type() == ir.Object {
				enc, err = option(enc, "encoding", constant("utf8"))
				if err != nil {
					return nil, true, err
				}
			}
			if err = encoding(enc); err != nil {
				return nil, true, err
			}
		}
		if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsStringLike == 0 {
			return nil, true, l.notYet(node, "node:fs.mkdtempSync outside the string UTF8 result")
		}
		args = args[:1]
	case "rm":
		recursive, force := ir.Expression(ir.BooleanConstant{Value: false}), ir.Expression(ir.BooleanConstant{Value: false})
		if len(args) > 1 {
			recursive, err = option(args[1], "recursive", recursive)
			if err != nil {
				return nil, true, err
			}
			force, err = option(args[1], "force", force)
			if err != nil {
				return nil, true, err
			}
			for _, name := range []string{"maxRetries", "retryDelay"} {
				fallback := number(0)
				if name == "retryDelay" {
					fallback = number(100)
				}
				value, e := option(args[1], name, fallback)
				if e != nil {
					return nil, true, e
				}
				if n, ok := value.(ir.NumberConstant); !ok || n.Value != fallback.(ir.NumberConstant).Value {
					return nil, true, l.notYet(node, "node:fs.rmSync retry options outside the driver defaults")
				}
			}
		}
		args = []ir.Expression{args[0], recursive, force}
	case "read_file":
		raw := l.nodeBufferType(l.checker.GetTypeAtLocation(node), "Buffer")
		if !raw && l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsStringLike == 0 {
			return nil, true, l.notYet(node, "node:fs.readFileSync with an ambiguous encoding")
		}
		flag := constant("r")
		enc := ir.Expression(ir.Undefined{Of: ir.String})
		if len(args) > 1 {
			enc = args[1]
		}
		if enc.Type() == ir.Object {
			flag, err = option(enc, "flag", flag)
			if err != nil {
				return nil, true, err
			}
			enc, err = option(enc, "encoding", constant("utf8"))
			if err != nil {
				return nil, true, err
			}
		}
		if err = encoding(enc); err != nil {
			return nil, true, err
		}
		args = []ir.Expression{args[0], flag}
		if raw {
			operation, of = "read_buffer", ir.Array
		}
	case "open":
		if args[1].Type() != ir.String {
			return nil, true, l.notYet(node, memberName+": "+"numeric fs open flags")
		}
		if len(args) < 3 {
			args = append(args, number(438))
		}
		if _, ok := args[2].(ir.Null); ok {
			args[2] = number(438)
		}
		if _, ok := args[2].(ir.Undefined); ok {
			args[2] = number(438)
		}
		if args[2].Type() != ir.Number {
			return nil, true, l.notYet(node, memberName+": "+"fs mode other than a number")
		}
	case "read_sync":
		if _, missing := args[2].(ir.Undefined); missing {
			args[2] = number(0)
		}
		if _, missing := args[4].(ir.Undefined); missing {
			args[4] = number(-1)
		}
		if _, null := args[4].(ir.Null); null {
			args[4] = number(-1)
		}
	case "write":
		position := number(-1)
		if len(args) > 2 {
			if _, ok := args[2].(ir.Undefined); !ok {
				if _, null := args[2].(ir.Null); !null {
					position = args[2]
				}
			}
		}
		if len(args) > 3 {
			if err = encoding(args[3]); err != nil {
				return nil, true, err
			}
		}
		if position.Type() != ir.Number {
			return nil, true, l.notYet(node, memberName+": "+"a union write position")
		}
		args = []ir.Expression{args[0], args[1], position}
	case "write_file":
		if args[1].Type() == ir.Array && l.nodeBufferType(l.checker.GetTypeAtLocation(call.Arguments.Nodes[1]), "Buffer") {
			operation = "write_buffer"
		}
		flag, mode, flush := constant("w"), number(438), ir.Expression(ir.BooleanConstant{Value: false})
		if len(args) > 2 {
			value := args[2]
			if value.Type() == ir.Object {
				if _, null := value.(ir.Null); !null {
					enc, e := option(value, "encoding", constant("utf8"))
					if e != nil {
						return nil, true, e
					}
					if e = encoding(enc); e != nil {
						return nil, true, e
					}
					flag, err = option(value, "flag", flag)
					if err != nil {
						return nil, true, err
					}
					mode, err = option(value, "mode", mode)
					if err != nil {
						return nil, true, err
					}
					flush, err = option(value, "flush", flush)
					if err != nil {
						return nil, true, err
					}
				}
			} else if err = encoding(value); err != nil {
				return nil, true, err
			}
		}
		args = []ir.Expression{args[0], args[1], flag, mode, flush}
	case "stat":
		throws := ir.Expression(ir.BooleanConstant{Value: true})
		if len(args) > 1 {
			throws, err = option(args[1], "throwIfNoEntry", throws)
			if err != nil {
				return nil, true, err
			}
			bigint, e := option(args[1], "bigint", ir.BooleanConstant{Value: false})
			if e != nil {
				return nil, true, e
			}
			if b, ok := bigint.(ir.BooleanConstant); !ok || b.Value {
				return nil, true, l.notYet(node, memberName+": "+"bigint fs Stats")
			}
		}
		args = []ir.Expression{args[0], throws}
	case "mkdir":
		recursive, mode := ir.Expression(ir.BooleanConstant{Value: false}), number(511)
		if len(args) > 1 {
			value := args[1]
			if _, null := value.(ir.Null); !null {
				if value.Type() == ir.Number {
					mode = value
				} else {
					recursive, err = option(value, "recursive", recursive)
					if err != nil {
						return nil, true, err
					}
					mode, err = option(value, "mode", mode)
					if err != nil {
						return nil, true, err
					}
				}
			}
		}
		args = []ir.Expression{args[0], recursive, mode}
	case "utimes":
		if args[1].Type() == ir.Object && args[2].Type() == ir.Object {
			operation = "utimes_dates"
		} else if args[1].Type() == ir.Object {
			operation = "utimes_atime_date"
		} else if args[2].Type() == ir.Object {
			operation = "utimes_mtime_date"
		}

	}
	// Separate descriptor operations preserve Node's ownership of caller fds.
	if operation == "read_file" && args[0].Type() == ir.Number {
		operation = "read_fd"
	}
	if operation == "read_buffer" && args[0].Type() == ir.Number {
		operation = "read_buffer_fd"
	}
	if operation == "write_buffer" && args[0].Type() == ir.Number {
		operation = "write_buffer_fd"
	}
	if operation == "write_file" && args[0].Type() == ir.Number {
		operation = "write_fd"
	}
	types := map[string][]ir.Type{
		"write_buffer": {ir.String, ir.Array, ir.String, ir.Number, ir.Boolean}, "write_buffer_fd": {ir.Number, ir.Array, ir.String, ir.Number, ir.Boolean},
		"read_sync": {ir.Number, ir.Array, ir.Number, ir.Number, ir.Number}, "read_buffer": {ir.String, ir.String}, "read_buffer_fd": {ir.Number, ir.String},
		"read_file": {ir.String, ir.String}, "read_fd": {ir.Number, ir.String}, "open": {ir.String, ir.String, ir.Number},
		"write": {ir.Number, ir.String, ir.Number}, "close": {ir.Number}, "write_file": {ir.String, ir.String, ir.String, ir.Number, ir.Boolean}, "write_fd": {ir.Number, ir.String, ir.String, ir.Number, ir.Boolean},
		"utimes_dates": {ir.String, ir.Object, ir.Object}, "utimes_atime_date": {ir.String, ir.Object, ir.Number}, "utimes_mtime_date": {ir.String, ir.Number, ir.Object},
		"mkdtemp": {ir.String}, "rm": {ir.String, ir.Boolean, ir.Boolean}, "exists": {ir.String}, "stat": {ir.String, ir.Boolean}, "mkdir": {ir.String, ir.Boolean, ir.Number}, "unlink": {ir.String}, "utimes": {ir.String, ir.Number, ir.Number},
	}
	for i, t := range types[operation] {
		if args[i].Type() != t {
			return nil, true, l.notYet(node, memberName+": "+"fs arguments outside the implemented host signatures")
		}
	}
	return ir.NodeFSFile{Operation: operation, Arguments: args, Of: of}, true, nil
}

func (l *lowering) nodeFSFileDate(node *ast.Node) (ir.Expression, bool, error) {
	created := node.AsNewExpression()
	if !l.isLibraryGlobal(created.Expression, "Date") {
		return nil, false, nil
	}
	if created.Arguments == nil || len(created.Arguments.Nodes) != 1 {
		return nil, true, l.notYet(node, "Date construction other than a numeric timestamp in the fs file host")
	}
	value, err := l.expression(created.Arguments.Nodes[0])
	if err != nil {
		return nil, true, err
	}
	if value.Type() != ir.Number {
		return nil, true, l.notYet(node, "Date construction other than a numeric timestamp in the fs file host")
	}
	return ir.NodeFSFile{Operation: "date_new", Arguments: []ir.Expression{value}, Of: ir.Object}, true, nil
}

// These native intrinsics inspect options synchronously and never write or
// retain them. Borrowing here does not authorize a wider mutable assignment.
func (l *lowering) nodeFSFileReadOnlyArgument(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent == nil || outer.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := outer.Parent.AsCallExpression()
	symbol := l.symbol(ast.SkipParentheses(call.Expression))
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	declaration := symbol.Declarations[0]
	if !load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) {
		return false
	}
	parent := declaration.Parent
	for parent != nil && parent.Kind != ast.KindModuleDeclaration {
		parent = parent.Parent
	}
	if parent == nil || (parent.Name().Text() != "node:fs" && parent.Name().Text() != "fs") {
		return false
	}
	index := 1
	switch symbol.Name {
	case "statSync", "mkdirSync", "readFileSync", "rmSync", "mkdtempSync":
	case "writeFileSync":
		index = 2
	default:
		return false
	}
	return len(call.Arguments.Nodes) > index && call.Arguments.Nodes[index] == outer
}

func init() {
	RegisterNodeLibraryMembers("node:fs.mkdtempSync", "node:fs.rmSync", "node:fs.readSync", "node:fs.readFileSync", "node:fs.openSync", "node:fs.writeSync", "node:fs.closeSync", "node:fs.writeFileSync", "node:fs.existsSync", "node:fs.statSync", "node:fs.mkdirSync", "node:fs.unlinkSync", "node:fs.utimesSync", "node:fs.StatsBase.isFile", "node:fs.StatsBase.isDirectory", "node:fs.StatsBase.isSymbolicLink", "node:fs.StatsBase.isBlockDevice", "node:fs.StatsBase.isCharacterDevice", "node:fs.StatsBase.isFIFO", "node:fs.StatsBase.isSocket", "node:fs.StatsBase.size", "node:fs.StatsBase.mtime", "node:fs.StatsBase.atime", "node:fs.StatsBase.mtimeMs", "node:globals.ErrnoException.code")
}
