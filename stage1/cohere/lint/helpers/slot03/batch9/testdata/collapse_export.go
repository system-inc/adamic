package tailwind

import "strings"

// Byte strings preserve exact private-helper argument identities across JSON.
func AdamicBytes(text string) string {
	r := make([]rune, len(text))
	for i, b := range []byte(text) {
		r[i] = rune(b)
	}
	return string(r)
}

type AdamicNode struct {
	Kind            string `json:"kind"`
	Selector        string `json:"selector"`
	Name            string `json:"name"`
	Params          string `json:"params"`
	Property        string `json:"property"`
	Value           string `json:"value"`
	Important       bool   `json:"important"`
	ValuePresent    bool   `json:"valuePresent"`
	ChildrenPresent bool   `json:"childrenPresent"`
}

func AdamicNodeOf(n *Node) AdamicNode {
	if n == nil {
		return AdamicNode{}
	}
	return AdamicNode{string(n.Kind), n.Selector, n.Name, n.Params, n.Property, n.Value, n.Important, n.ValuePresent, n.Nodes != nil}
}

type AdamicDeclaration struct {
	Present bool       `json:"present"`
	Node    AdamicNode `json:"node"`
}
type AdamicString struct {
	End     int    `json:"end"`
	Failed  bool   `json:"failed"`
	Message string `json:"message"`
	Offset  int    `json:"offset"`
}
type AdamicDependencies struct {
	AtRules      map[string]AdamicNode        `json:"atRules"`
	Declarations map[string]AdamicDeclaration `json:"declarations"`
	Strings      map[string]AdamicString      `json:"strings"`
	Trims        map[string]string            `json:"trims"`
}

var AdamicCalls AdamicDependencies

func AdamicReset() {
	AdamicCalls = AdamicDependencies{map[string]AdamicNode{}, map[string]AdamicDeclaration{}, map[string]AdamicString{}, map[string]string{}}
}
func adamicTrim(text string) string {
	out := strings.TrimSpace(text)
	AdamicCalls.Trims[AdamicBytes(text)] = AdamicBytes(out)
	return out
}
func ParseAtRule(buffer string, nodes ...*Node) *Node {
	n := adamicOriginalAtRule(buffer, nodes...)
	AdamicCalls.AtRules[AdamicBytes(buffer)] = AdamicNodeOf(n)
	return n
}
func parseCSSDeclaration(buffer string, colon int) *Node {
	n := adamicOriginalDeclaration(buffer, colon)
	AdamicCalls.Declarations[AdamicBytes(buffer)+"|"+adamicInteger(colon)] = AdamicDeclaration{n != nil, AdamicNodeOf(n)}
	return n
}
func parseCSSString(input string, start int, quote byte) (int, error) {
	end, err := adamicOriginalString(input, start, quote)
	v := AdamicString{End: end}
	if err != nil {
		e := err.(*CSSSyntaxError)
		v.Failed = true
		v.Message = e.Message
		v.Offset = e.Offset
	}
	AdamicCalls.Strings[adamicInteger(start)+"|"+adamicInteger(int(quote))] = v
	return end, err
}
func AdamicThemeHandle(handle int) (int, bool, bool) {
	themes := []*Theme{NewTheme(), NewTheme()}
	var theme *Theme
	if handle >= 0 {
		theme = themes[handle]
	}
	system := &LoadedDesignSystem{theme: theme}
	got := system.Theme()
	id := -1
	for i, t := range themes {
		if got == t {
			id = i
		}
	}
	alias := got == theme
	unchanged := system.theme == theme
	if got != nil {
		got.Prefix = "mutated"
		alias = alias && theme.Prefix == "mutated"
	}
	return id, alias, unchanged
}
