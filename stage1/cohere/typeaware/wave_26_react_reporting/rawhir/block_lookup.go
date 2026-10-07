package high_level_intermediate_representation

func blockById(function *Function, id BlockId) *BasicBlock {
	for _, block := range function.Blocks {
		if block != nil && block.Id == id {
			return block
		}
	}
	return nil
}
