package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Shrink makes a failing program as small as it can while it still fails the same way (the same
// Key), by delta debugging: whole runs of statements removed, a compound statement replaced by what
// its blocks hold, and an expression replaced by a literal of its type or by one of its own children
// of the same type. It repeats until nothing more goes, or its budget of tries is spent.
//
// Every candidate is a whole program run three ways, so a candidate the checker refuses, or one
// that now fails some other way, is simply not kept.
func Shrink(program *Program, key string, try func(*Program) Outcome) *Program {
	shrinker := &shrinker{current: program.Clone(), key: key, try: try, seen: map[string]bool{}, budget: 3000}
	for shrinker.budget > 0 {
		before := shrinker.current.Source()
		for _, block := range shrinker.blocks() {
			shrinker.statements(block)
		}
		for _, block := range shrinker.blocks() {
			shrinker.inline(block)
		}
		shrinker.expressions()
		if shrinker.current.Source() == before {
			break
		}
	}
	return shrinker.current
}

type shrinker struct {
	current *Program
	key     string
	try     func(*Program) Outcome
	// seen holds every source tried, so the same candidate never runs twice.
	seen   map[string]bool
	budget int
}

// fails says whether the current program, as it now stands, still fails the same way.
func (s *shrinker) fails() bool {
	source := s.current.Source()
	// A move refusal's comments alone would make any accepted empty program
	// look like the same compiler regression. Keep the actual task boundary.
	if s.current.RefusalFix != "" && !strings.Contains(source, "parallelMap(") {
		return false
	}
	if s.seen[source] || s.budget <= 0 {
		return false
	}
	s.seen[source] = true
	s.budget--
	outcome := s.try(s.current)
	return outcome.Verdict == Finding && outcome.Key == s.key
}

// blocks is every block in the program, outermost first.
func (s *shrinker) blocks() []*Block {
	blocks := []*Block{&s.current.Block}
	for index := 0; index < len(blocks); index++ {
		for _, statement := range blocks[index].Statements {
			for _, part := range statement.Parts {
				if part.Block != nil {
					blocks = append(blocks, part.Block)
				}
			}
		}
	}
	return blocks
}

// statements removes as many of a block's statements as it can, by ddmin: the block split in n
// chunks, each chunk's removal tried, n doubled when none can go.
func (s *shrinker) statements(block *Block) {
	chunks := 2
	for len(block.Statements) > 0 && s.budget > 0 {
		kept := block.Statements
		size := (len(kept) + chunks - 1) / chunks
		removed := false
		for start := 0; start < len(kept); start += size {
			end := min(start+size, len(kept))
			block.Statements = append(append([]*Statement{}, kept[:start]...), kept[end:]...)
			if s.fails() {
				removed = true
				break
			}
			block.Statements = kept
		}
		switch {
		case removed:
			chunks = max(chunks-1, 2)
		case size == 1:
			return
		default:
			chunks = min(chunks*2, len(kept))
		}
	}
}

// inline replaces a statement that holds blocks (an if, a loop, a function) with the statements of
// one of its blocks, which keeps what the block did and drops the condition or the loop around it.
func (s *shrinker) inline(block *Block) {
	for index := 0; index < len(block.Statements); index++ {
		statement := block.Statements[index]
		for _, part := range statement.Parts {
			if part.Block == nil {
				continue
			}
			kept := block.Statements
			replaced := append(append(append([]*Statement{}, kept[:index]...), part.Block.Statements...), kept[index+1:]...)
			block.Statements = replaced
			if s.fails() {
				break
			}
			block.Statements = kept
		}
	}
}

// expressions simplifies every expression: replaced by a literal of its type, or by one of its own
// children of the same type.
func (s *shrinker) expressions() {
	for _, slot := range s.slots() {
		original := *slot
		if original.Type == Other || isLiteral(original) {
			continue
		}
		var candidates []*Expression
		candidates = append(candidates, simplest(original.Type))
		for _, part := range original.Parts {
			if part.Expression != nil && part.Expression.Type == original.Type {
				candidates = append(candidates, part.Expression)
			}
		}
		for _, candidate := range candidates {
			if candidate == nil {
				continue
			}
			*slot = candidate
			if s.fails() {
				break
			}
			*slot = original
		}
	}
}

// slots is a pointer to every expression in the program, parents before children, so replacing a
// parent first spares trying each of its children.
func (s *shrinker) slots() []**Expression {
	var slots []**Expression
	var visit func(slot **Expression)
	visit = func(slot **Expression) {
		slots = append(slots, slot)
		for index := range (*slot).Parts {
			if (*slot).Parts[index].Expression != nil {
				visit(&(*slot).Parts[index].Expression)
			}
		}
	}
	for _, block := range s.blocks() {
		for _, statement := range block.Statements {
			for index := range statement.Parts {
				if statement.Parts[index].Expression != nil {
					visit(&statement.Parts[index].Expression)
				}
			}
		}
	}
	return slots
}

// simplest is the plainest literal of a type, or nil where there's none worth writing.
func simplest(t Type) *Expression {
	switch t {
	case Number:
		return text(Number, "0")
	case String:
		return text(String, "''")
	case Boolean:
		return text(Boolean, "true")
	case NumberArray:
		return text(NumberArray, "[0]")
	case StringArray:
		return text(StringArray, "['']")
	}
	return nil
}

// WriteFinding writes a shrunk finding as a program with a header that says where it came from and
// how to make it again.
func WriteFinding(path string, program *Program, seed uint64, without []string, key string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	command := fmt.Sprintf("go run ./cmd/adamic-fuzz -seed %d -count 1", seed)
	if program.Feature == "moves" {
		command += " -only-moves"
	}
	if len(without) > 0 {
		command += " -without " + strings.Join(without, ",")
	}
	header := fmt.Sprintf("// Found by adamic-fuzz from seed %d (%s), shrunk: %s.\n", seed, command, strings.ReplaceAll(key, "\n", " "))
	return os.WriteFile(path, []byte(header+program.Source()), 0o644)
}
