package parser

const testJsxMutantsShards = 9

func jsxMutantsChanges() []struct{ name, from, to string } {
	return []struct{ name, from, to string }{
		{"text payload", "this.parser.node(id).text = this.parser.scanner.value;", "this.parser.node(id).text = this.parser.scanner.value + '!';"},
		{"whitespace flag", "whitespace ? '1' : '0'", "whitespace ? '0' : '0'"},
		{"namespace kind", "this.make('JsxNamespacedName', pos, [name, right])", "this.make('QualifiedName', pos, [name, right])"},
		{"self closing kind", "this.make('JsxSelfClosingElement', pos, header)", "this.make('JsxOpeningElement', pos, header)"},
		{"type argument comma", "this.parser.node(id).semantic = `${typeCount}:${typeTrailing ? 1 : 0}`;", "this.parser.node(id).semantic = `${typeCount}:0`;"},
		{"attribute list", "this.parser.node(id).list = attributes.length;", "this.parser.node(id).list = attributes.length + 1;"},
		{"expression kind", "this.make('JsxExpression', pos, children)", "this.make('ParenthesizedExpression', pos, children)"},
		{"child order", "this.make(fragment ? 'JsxFragment' : 'JsxElement', pos, children)", "this.make(fragment ? 'JsxFragment' : 'JsxElement', pos, children.slice().reverse())"},
		{"text start", "const id = this.make('JsxText', pos);", "const id = this.make('JsxText', pos + 1);"},
	}
}
