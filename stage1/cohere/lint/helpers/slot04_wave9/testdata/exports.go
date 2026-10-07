package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

type AdamicUtilityInput struct {
	Params  string
	Body    []int
	Trimmed string
}
type AdamicNormalizationNode struct {
	Tag, Text string
	Edges     []int
}
type AdamicAdaptation struct {
	Nodes     []AdamicNormalizationNode
	Roots     []int
	Segments  map[string][]string
	Arguments map[string]string
	Parses    map[string][]int
}

func AdamicUtility(inputs []AdamicUtilityInput, path string, keys []string) {
	c := &stylesheetCollector{utilityRoots: map[string]map[UtilityKind]bool{}, staticUtilityNodes: map[string][]*Node{}}
	ids := map[*Node]int{}
	for _, in := range inputs {
		nodes := []*Node{}
		for _, id := range in.Body {
			n := &Node{Kind: KindComment}
			ids[n] = id
			nodes = append(nodes, n)
		}
		err := c.ingestUtilityBlock(&Node{Params: in.Params, Nodes: nodes}, path)
		message := ""
		if err != nil {
			message = err.Error()
		}
		fmt.Println(message)
		states := []string{}
		for _, key := range keys {
			bits := 0
			if c.utilityRoots[key][UtilityKindStatic] {
				bits++
			}
			if c.utilityRoots[key][UtilityKindFunctional] {
				bits += 2
			}
			body := []string{}
			present := false
			if ns, ok := c.staticUtilityNodes[key]; ok {
				present = true
				for _, n := range ns {
					body = append(body, fmt.Sprint(ids[n]))
				}
			}
			states = append(states, fmt.Sprintf("%s:%d:%t:%s", key, bits, present, strings.Join(body, ",")))
		}
		defs := []string{}
		for _, d := range c.utilityDefinitions {
			body := []string{}
			for _, n := range d.Nodes {
				body = append(body, fmt.Sprint(ids[n]))
			}
			defs = append(defs, d.Name+":"+strings.Join(body, ","))
		}
		fmt.Println(strings.Join(states, "|") + ";" + strings.Join(defs, "|"))
	}
}
func AdamicAdapt(source string) AdamicAdaptation {
	a := AdamicAdaptation{Nodes: []AdamicNormalizationNode{}, Roots: []int{}, Segments: map[string][]string{}, Arguments: map[string]string{}, Parses: map[string][]int{}}
	var flatten func([]ValueNode) []int
	flatten = func(nodes []ValueNode) []int {
		roots := []int{}
		for _, n := range nodes {
			id := len(a.Nodes)
			a.Nodes = append(a.Nodes, AdamicNormalizationNode{string(n.Kind), n.Value, []int{}})
			a.Nodes[id].Edges = flatten(n.Nodes)
			roots = append(roots, id)
		}
		return roots
	}
	nodes := ParseValue(source)
	a.Roots = flatten(nodes)
	var collect func([]ValueNode)
	collect = func(nodes []ValueNode) {
		for _, n := range nodes {
			if n.Kind != ValueNodeKindFunction {
				continue
			}
			if n.Value != "--value" && n.Value != "--modifier" {
				collect(n.Nodes)
				continue
			}
			s := ValueToCss(n.Nodes)
			args := segment(s, ',')
			a.Segments[s] = append([]string{}, args...)
			normalized := []string{}
			for _, arg := range args {
				v := normalizeValueFunctionArgument(arg)
				a.Arguments[arg] = v
				normalized = append(normalized, v)
			}
			joined := strings.Join(normalized, ",")
			if _, ok := a.Parses[joined]; !ok {
				a.Parses[joined] = flatten(ParseValue(joined))
			}
		}
	}
	collect(nodes)
	return a
}
func AdamicNormalize(source string) {
	nodes := ParseValue(source)
	normalizeValueFunctionNodes(nodes)
	fmt.Println(ValueToCss(nodes))
}

var adamicRecordLock sync.Mutex

func adamicRecordCase(row any) {
	path := os.Getenv("ADAMIC_SLOT04_WAVE9")
	if path == "" {
		return
	}
	adamicRecordLock.Lock()
	defer adamicRecordLock.Unlock()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(row); err != nil {
		panic(err)
	}
}
func AdamicRecordUtility(node *Node, path string) {
	body := []int{}
	for i := range node.Nodes {
		body = append(body, i)
	}
	name := strings.TrimSpace(node.Params)
	root := strings.TrimSuffix(name, "-*")
	adamicRecordCase(map[string]any{"Kind": "utility", "Path": path, "Inputs": []AdamicUtilityInput{{Params: node.Params, Body: body}}, "Keys": []string{root}})
}
func AdamicRecordNormalization(nodes []ValueNode) {
	adamicRecordCase(map[string]any{"Kind": "normalize", "Source": ValueToCss(nodes)})
}
