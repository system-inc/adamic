package tailwind

func AdamicHoleEdges(before, after string, first, last, leading, trailing bool) (bool, bool) {
	edges := holeEdges(before, after, first, last, classValueEdges{Leading: leading, Trailing: trailing})
	return edges.Leading, edges.Trailing
}
