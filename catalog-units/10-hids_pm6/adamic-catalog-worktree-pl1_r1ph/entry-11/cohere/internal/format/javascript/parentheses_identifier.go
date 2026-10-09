package javascript

import "regexp"

// parentheses/identifier.js.

var htmlPlaceholderPattern = regexp.MustCompile(`^PRETTIER_HTML_PLACEHOLDER_\d+_\d+_IN_JS$`)

func shouldAddParenthesesToIdentifier(path *Path) bool {
	current := node(path)

	// Identifiers never need parentheses.
	if !current.Is("Identifier") {
		return false
	}

	// ...unless those identifiers are embed placeholders. They might be substituted by complex
	// expressions, so the parens around them should not be dropped. Example (JS-in-HTML-in-JS):
	//     let tpl = html`<script> f((${expr}) / 2); </script>`;
	// If the inner JS formatter removes the parens, the expression might change its meaning:
	//     f((a + b) / 2)  vs  f(a + b / 2)
	if current.Parenthesized &&
		htmlPlaceholderPattern.MatchString(current.String("name")) {
		return true
	}

	key := keyOf(path)
	parent := parentOf(path)
	name := current.String("name")
	// `for ((async) of []);` and `for ((let) of []);`
	if key == "left" &&
		(name == "async" && !parent.Truthy("await") || name == "let") &&
		parent.Is("ForOfStatement") {
		return true
	}

	// `for ((let.a) of []);`
	// `for ((let.a) in []);`
	if name == "let" {
		ancestor, _ := path.FindAncestor(func(node Node) bool {
			return node.Is("ForOfStatement") || node.Is("ForInStatement")
		})
		expression := ancestor.Child("left")
		if expression != nil &&
			startsWithNoLookaheadToken(expression, func(leftmostNode Node) bool { return leftmostNode == current }) {
			return true
		}
	}

	// `(let)[a] = 1`
	if key == "object" &&
		name == "let" &&
		parent.Is("MemberExpression") &&
		parent.Truthy("computed") &&
		!parent.Truthy("optional") {
		statement, _ := path.FindAncestor(func(node Node) bool {
			return node.Is("ExpressionStatement") ||
				node.Is("ForStatement") ||
				node.Is("ForInStatement")
		})
		// Upstream's nested ternary over statement.type.
		var expression Node
		switch {
		case statement == nil:
			expression = nil
		case statement.Is("ExpressionStatement"):
			expression = statement.Child("expression")
		case statement.Is("ForStatement"):
			expression = statement.Child("init")
		default:
			expression = statement.Child("left")
		}
		if expression != nil &&
			startsWithNoLookaheadToken(expression, func(leftmostNode Node) bool { return leftmostNode == current }) {
			return true
		}
	}

	// `(type) satisfies never;` and similar cases
	if key == "expression" {
		switch name {
		case "await", "interface", "module", "using", "yield", "let", "component", "hook", "type":
			ancestorNeitherAsNorSatisfies, _ := path.FindAncestor(func(node Node) bool {
				return !isBinaryCastExpression(node)
			})
			if ancestorNeitherAsNorSatisfies != parent &&
				ancestorNeitherAsNorSatisfies.Is("ExpressionStatement") {
				return true
			}
		}
	}

	return false
}
