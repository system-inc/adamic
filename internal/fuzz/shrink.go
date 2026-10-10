package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Shrink makes a failing program as small as it can while it still fails the same way (the same
// Key), by delta debugging: whole runs of statements removed, a compound statement replaced by what
// its blocks hold, and an expression replaced by a literal of its type or by one of its own children
// of the same type. It repeats until nothing more goes, or its budget of tries is spent.
//
// Every candidate is a whole program run three ways, so a candidate the checker refuses, or one
// that now fails some other way, is simply not kept. try runs one candidate at a time.
func Shrink(program *Program, key string, try func(*Program) Outcome) *Program {
	reduced := reduce(&programTree{program: program.Clone()}, func(candidate Reducible) bool {
		// A move refusal's comments alone would make any accepted empty program
		// look like the same compiler regression. Keep the actual task boundary.
		if program.RefusalFix != "" && !strings.Contains(candidate.Source(), "parallelMap(") {
			return false
		}
		outcome := try(candidate.(*programTree).program)
		return outcome.Verdict == Finding && outcome.Key == key
	}, 1, 3000)
	return reduced.(*programTree).program
}

// Reducible is a program the shrinker can make smaller: lists whose items can go, compounds that can
// be replaced by what one of their parts holds, and expressions that can be replaced by simpler ones.
// Each is named by its index, which holds only for the value it was asked of; a candidate is always a
// new value, never a change to this one, so several can be tried at once.
type Reducible interface {
	Source() string
	// Lists counts the lists of removable items, outermost first, so taking items from one list
	// leaves the lists before it, and its own index, where they were.
	Lists() int
	// Items counts a list's items.
	Items(list int) int
	// Without is a copy with a list's items from start up to end taken out.
	Without(list, start, end int) Reducible
	// Compounds counts the items that hold lists of their own.
	Compounds() int
	// Inlined is every copy with one compound replaced by what one of its parts holds.
	Inlined(compound int) []Reducible
	// Expressions counts the expressions, parents before children.
	Expressions() int
	// Simpler is every copy with one expression replaced by a simpler one of the same type.
	Simpler(expression int) []Reducible
}

// reduce makes start as small as it can while keep holds, trying up to workers candidates at once,
// until a whole round changes nothing or budget candidates have been tried. When several candidates
// at once are kept, the first in order wins, so the result is the same at any number of workers.
func reduce(start Reducible, keep func(Reducible) bool, workers int, budget int) Reducible {
	shrinker := &shrinker{current: start, keep: keep, workers: max(workers, 1), budget: budget, seen: map[string]bool{start.Source(): true}}
	for shrinker.budget > 0 {
		before := shrinker.current.Source()
		for list := 0; list < shrinker.current.Lists(); list++ {
			shrinker.items(list)
		}
		shrinker.each(func() int { return shrinker.current.Compounds() }, func(index int) []Reducible { return shrinker.current.Inlined(index) })
		shrinker.each(func() int { return shrinker.current.Expressions() }, func(index int) []Reducible { return shrinker.current.Simpler(index) })
		if shrinker.current.Source() == before {
			break
		}
	}
	return shrinker.current
}

type shrinker struct {
	current Reducible
	keep    func(Reducible) bool
	workers int
	// seen holds every source tried, so the same candidate never runs twice.
	seen   map[string]bool
	budget int
}

// first tries candidates in order, workers at a time, and makes the first one kept current. It
// says which, or -1 when none was.
func (s *shrinker) first(candidates []Reducible) int {
	for start := 0; start < len(candidates) && s.budget > 0; start += s.workers {
		var batch []int
		for index := start; index < min(start+s.workers, len(candidates)) && len(batch) < s.budget; index++ {
			source := candidates[index].Source()
			if s.seen[source] {
				continue
			}
			s.seen[source] = true
			batch = append(batch, index)
		}
		s.budget -= len(batch)
		kept := make([]bool, len(batch))
		var group sync.WaitGroup
		for slot, index := range batch {
			group.Add(1)
			go func() {
				defer group.Done()
				kept[slot] = s.keep(candidates[index])
			}()
		}
		group.Wait()
		for slot, index := range batch {
			if kept[slot] {
				s.current = candidates[index]
				return index
			}
		}
	}
	return -1
}

// items removes as many of a list's items as it can, by ddmin: the list split in n chunks, each
// chunk's removal tried, n doubled when none can go.
func (s *shrinker) items(list int) {
	chunks := 2
	for list < s.current.Lists() && s.current.Items(list) > 0 && s.budget > 0 {
		count := s.current.Items(list)
		size := (count + chunks - 1) / chunks
		var candidates []Reducible
		for start := 0; start < count; start += size {
			candidates = append(candidates, s.current.Without(list, start, min(start+size, count)))
		}
		switch {
		case s.first(candidates) >= 0:
			chunks = max(chunks-1, 2)
		case size == 1:
			return
		default:
			chunks = min(chunks*2, count)
		}
	}
}

// each tries the candidates for every index counted, in order, workers at a time across indexes, and
// goes on after the one kept: what it replaced is gone, and the indexes after it are counted again.
func (s *shrinker) each(count func() int, candidates func(int) []Reducible) {
	for index := 0; index < count() && s.budget > 0; {
		var batch []Reducible
		var owners []int
		next := index
		for ; next < count() && len(batch) < s.workers; next++ {
			for _, candidate := range candidates(next) {
				batch = append(batch, candidate)
				owners = append(owners, next)
			}
		}
		kept := s.first(batch)
		if kept < 0 {
			index = next
			continue
		}
		index = owners[kept] + 1
	}
}

// programTree is a generated program as the shrinker sees it: the lists are its blocks, the
// compounds the statements that hold blocks, and the expressions every one but a literal or a value
// of a type the generator doesn't substitute.
type programTree struct {
	program *Program
}

func (t *programTree) Source() string { return t.program.Source() }

func (t *programTree) Lists() int { return len(blocksOf(t.program)) }

func (t *programTree) Items(list int) int { return len(blocksOf(t.program)[list].Statements) }

func (t *programTree) Without(list, start, end int) Reducible {
	copied := t.program.Clone()
	block := blocksOf(copied)[list]
	block.Statements = append(append([]*Statement{}, block.Statements[:start]...), block.Statements[end:]...)
	return &programTree{program: copied}
}

// compound is a statement that holds blocks, by where it is.
type compound struct {
	block     int
	statement int
}

func compoundsOf(program *Program) []compound {
	var found []compound
	for blockIndex, block := range blocksOf(program) {
		for statementIndex, statement := range block.Statements {
			for _, part := range statement.Parts {
				if part.Block != nil {
					found = append(found, compound{block: blockIndex, statement: statementIndex})
					break
				}
			}
		}
	}
	return found
}

func (t *programTree) Compounds() int { return len(compoundsOf(t.program)) }

func (t *programTree) Inlined(index int) []Reducible {
	where := compoundsOf(t.program)[index]
	var candidates []Reducible
	for part := range blocksOf(t.program)[where.block].Statements[where.statement].Parts {
		copied := t.program.Clone()
		block := blocksOf(copied)[where.block]
		statement := block.Statements[where.statement]
		if statement.Parts[part].Block == nil {
			continue
		}
		block.Statements = append(append(append([]*Statement{}, block.Statements[:where.statement]...), statement.Parts[part].Block.Statements...), block.Statements[where.statement+1:]...)
		candidates = append(candidates, &programTree{program: copied})
	}
	return candidates
}

func (t *programTree) Expressions() int { return len(slotsOf(t.program)) }

func (t *programTree) Simpler(index int) []Reducible {
	original := *slotsOf(t.program)[index]
	if original.Type == Other || isLiteral(original) {
		return nil
	}
	replacements := []*Expression{simplest(original.Type)}
	for _, part := range original.Parts {
		if part.Expression != nil && part.Expression.Type == original.Type {
			replacements = append(replacements, part.Expression)
		}
	}
	var candidates []Reducible
	for which, replacement := range replacements {
		if replacement == nil {
			continue
		}
		copied := t.program.Clone()
		slot := slotsOf(copied)[index]
		if which == 0 {
			*slot = replacement
		} else {
			// The same child, in the copy.
			for _, part := range (*slot).Parts {
				if part.Expression != nil && part.Expression.Type == original.Type {
					which--
					if which == 0 {
						*slot = part.Expression
						break
					}
				}
			}
		}
		candidates = append(candidates, &programTree{program: copied})
	}
	return candidates
}

// blocksOf is every block in the program, outermost first.
func blocksOf(program *Program) []*Block {
	blocks := []*Block{&program.Block}
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

// slotsOf is a pointer to every expression in the program, parents before children, so replacing a
// parent first spares trying each of its children.
func slotsOf(program *Program) []**Expression {
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
	for _, block := range blocksOf(program) {
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
func WriteFinding(path string, program *Program, seed uint64, without []string, with []string, key string) error {
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
	if len(with) > 0 {
		command += " -with " + strings.Join(with, ",")
	}
	header := fmt.Sprintf("// Found by adamic-fuzz from seed %d (%s), shrunk: %s.\n", seed, command, strings.ReplaceAll(key, "\n", " "))
	return os.WriteFile(path, []byte(header+program.Source()), 0o644)
}
