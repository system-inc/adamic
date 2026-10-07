package checker

// The nested payload uses the existing raw type graph framing.
func (p *Program) wave19ContractTypes(g *graph, roots []uint64) string {
	out := &fields{}
	out.number(1)
	out.text("wave19-contract-types")
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(true)
	var nonzero []uint64
	for _, id := range roots {
		if id != 0 {
			nonzero = append(nonzero, id)
		}
	}
	out.ids(nonzero)
	out.number(0)
	g.write(out)
	return out.String()
}
