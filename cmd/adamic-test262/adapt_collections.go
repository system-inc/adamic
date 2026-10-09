package main

import "strings"

const adaptCollectionType = "collection-type"

func (p *parser) noteCollectionCall(callee *texpr, args []*texpr) {
	if callee != nil && callee.op == "member" {
		p.collectionCalls = append(p.collectionCalls, callNote{method: callee.prop, recv: callee.left, args: args})
	}
}

// Empty constructors do not infer their generic arguments from subsequent writes in TypeScript.
// Supply only erased type arguments, after proving every reference is a direct library operation
// and every argument has one agreed primitive type. No aliases, escapes or reassignments qualify.
func (p *parser) annotateCollections() {
	if p.unsafeVars {
		return
	}
	for _, stmt := range p.varStmts {
		for _, decl := range stmt.decls {
			b := decl.binding
			if !p.declSafe(decl) || b.init == nil || b.init.op != "emptyCollection" || len(b.assigns) != 0 {
				continue
			}
			ctor := b.init
			if lookup(ctor.scope, ctor.name) != nil {
				continue
			}
			safe := true
			for _, ref := range p.refs {
				if lookup(ref.scope, ref.name) != b {
					continue
				}
				i := ref.at
				if i < 0 || i+2 >= len(p.toks) || p.toks[i+1].text != "." {
					safe = false
					break
				}
				method := p.toks[i+2].text
				if method == "size" {
					if i+3 < len(p.toks) && (p.toks[i+3].text == "=" || strings.HasSuffix(p.toks[i+3].text, "=")) {
						safe = false
					}
					continue
				}
				if i+3 >= len(p.toks) || p.toks[i+3].text != "(" {
					safe = false
					break
				}
				if method != "has" && method != "delete" && method != "clear" && !(ctor.name == "Set" && method == "add") && !(ctor.name == "Map" && (method == "get" || method == "set")) {
					safe = false
					break
				}
			}
			var key, value *jtype
			for _, call := range p.collectionCalls {
				if call.recv == nil || call.recv.op != "name" || lookup(call.recv.scope, call.recv.name) != b {
					continue
				}
				n := 1
				if call.method == "clear" {
					n = 0
				}
				if call.method == "set" {
					n = 2
				}
				if len(call.args) != n {
					safe = false
					break
				}
				for i, arg := range call.args {
					typ := p.eval(arg)
					if typ == nil || (typ.kind != "number" && typ.kind != "string" && typ.kind != "boolean") {
						safe = false
						break
					}
					chosen := &key
					if i == 1 {
						chosen = &value
					}
					if *chosen == nil {
						*chosen = typ
					} else if !sameType(*chosen, typ) {
						safe = false
						break
					}
				}
			}
			if !safe || key == nil || (ctor.name == "Map" && value == nil) {
				continue
			}
			types := key.string()
			if ctor.name == "Map" {
				types += ", " + value.string()
			}
			p.edits = append(p.edits, edit{start: ctor.constructorEnd, end: ctor.constructorEnd, text: "<" + types + ">"})
			p.counts[adaptCollectionType]++
		}
	}
}
