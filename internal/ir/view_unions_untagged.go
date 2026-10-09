package ir

// ViewUnionHasDiscriminant requires a shared required field whose finite literal
// sets select exactly one member. A common enum field with overlapping values
// does not provide a member certificate.
func ViewUnionHasDiscriminant(contracts []ViewContract, root ViewContract) bool {
	if root.Kind != ViewUnion || len(root.Members) == 0 {
		return false
	}
	for _, field := range root.Fields {
		if field.Optional {
			continue
		}
		seen := map[ViewLiteral]bool{}
		valid := true
		for _, id := range root.Members {
			if id <= 0 || int(id) > len(contracts) {
				valid = false
				break
			}
			member := contracts[id-1]
			found := false
			for _, own := range member.Fields {
				if own.Name != field.Name || own.Optional || own.Contract <= 0 || int(own.Contract) > len(contracts) {
					continue
				}
				tag := contracts[own.Contract-1]
				if tag.Kind != ViewScalar || len(tag.Allowed) == 0 {
					break
				}
				found = true
				for _, literal := range tag.Allowed {
					if (literal.Of != Number && literal.Of != String && literal.Of != Boolean) || seen[literal] {
						valid = false
						break
					}
					seen[literal] = true
				}
				break
			}
			if !found {
				valid = false
			}
			if !valid {
				break
			}
		}
		if valid {
			return true
		}
	}
	return false
}
