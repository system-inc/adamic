package tailwind

func AdamicNextOrder(groupPresent bool, groupOrder, lastOrder int) (int, bool, int, int, int) {
	registry := NewVariantRegistry()
	registry.lastOrder = lastOrder
	if groupPresent {
		registry.groupOrder = &groupOrder
	}
	first := registry.nextOrder()
	present := registry.groupOrder != nil
	group := groupOrder
	last := registry.lastOrder
	second := registry.nextOrder()
	return first, present, group, last, second
}
