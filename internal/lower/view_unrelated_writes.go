package lower

import "github.com/system-inc/adamic/internal/ir"

// Only a closed may-flow proof can distinguish an ordinary reference write from
// a checked write sharing its field name. Unknown receivers retain the check.
func (l *lowering) certifyUnrelatedViewWrites(graph *allocationFlowGraph, viewed map[int]bool, unknown bool) {
	if unknown {
		return
	}
	rewrite := func(node any) any {
		if read, ok := node.(ir.Property); ok && read.Of.IsReference() && read.Of != ir.Closure && read.View != "" {
			reaches := graph.ReachingAllocations(read.Object)
			unrelated := !reaches.Unknown && len(reaches.Sites) != 0
			for _, site := range reaches.Sites {
				unrelated = unrelated && !viewed[site]
			}
			if unrelated {
				read.ViewOrdinary = true
				return read
			}
		}
		write, ok := node.(ir.SetProperty)
		if !ok || write.Value == nil || !write.Value.Type().IsReference() || !l.result.CheckedFields[write.Name] {
			return node
		}
		reaches := graph.ReachingAllocations(write.Object)
		if reaches.Unknown || len(reaches.Sites) == 0 {
			return node
		}
		for _, site := range reaches.Sites {
			if viewed[site] {
				return node
			}
		}
		write.ViewWriteUnrelated = true
		return write
	}
	l.result.Main = rewriteShapeStatements(l.result.Main, rewrite)
	for i := range l.result.Functions {
		l.result.Functions[i].Body = rewriteShapeStatements(l.result.Functions[i].Body, rewrite)
	}
}
