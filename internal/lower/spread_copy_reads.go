package lower

import (
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// A spread copies own properties only, and a class's getter is its prototype's, so { ...instance } holds
// no field of that getter's name, while the checker types the copy with one when the source is seen as a
// type that declares it (#7mjgxkc). A read of the name from the copy loaded a slot the copy doesn't have:
// natively garbage where Node prints undefined (c13b), or a stop on a compiler bug (c13c). Compiler's
// ruling (B) on #7mjgxkc stops the read, not the spread, so a copy whose getter's name nothing reads
// lowers as on main. The stop runs in finishAccessors before the dispatcher or the plain load, the two
// roads that would load the slot, so neither is reached for such a read: a receiver whose allocations
// may hold such a copy stops, and so does one whose reach the graph can't close while any spread in the
// program may copy such a getter, failing closed. An accessor literal's getter is its own, which a spread
// runs and copies the value of, so only a class's counts. The real fix, the copy's read answering Node's
// undefined, is the object model's.

// spreadCopiedGetters is every getter name some spread in the program may leave out of its copy, each with
// the first class found holding it: a getter spreadGetters says the spread may meet on a class's chain,
// not an accessor literal's.
//
// A literal that writes such a name after its spread stops here: natively the copy is laid out on the
// source class's data shape, which has no slot for the getter's name, so the write is lost, the read after
// it stopped on a compiler bug and an enumeration of the copy would leave the field out where Node holds it.
func (l *lowering) spreadCopiedGetters() (map[string]string, error) {
	copied := map[string]string{}
	l.spreadWithCopied = map[string]string{}
	var failure error
	check := func(node any) bool {
		if literal, ok := node.(ir.ObjectLiteral); ok && literal.Spread != nil && failure == nil {
			for name, class := range l.spreadCopyGetters(literal) {
				if slices.ContainsFunc(literal.Fields, func(field ir.Field) bool { return field.Name == name }) {
					failure = &NotYet{Where: l.result.Source, What: "writing " + name + " after the spread in " + literal.SpreadReadiness + ", a spread's copy of a " + class + " instance, whose " + name + " is a getter on its prototype: the copy has no slot for " + name}
					return false
				}
				if _, known := copied[name]; !known {
					copied[name] = class
				}
				if _, known := l.spreadWithCopied[name]; literal.SpreadWith && !known {
					l.spreadWithCopied[name] = class
				}
			}
		}
		return true
	}
	walk(l.result.Main, check)
	for _, function := range l.result.Functions {
		walk(function.Body, check)
	}
	return copied, failure
}

// spreadCopyGetters is, for one spread's literal, each getter name it may leave out of its copy with the
// class holding it (spreadCopiedGetters' rule).
func (l *lowering) spreadCopyGetters(literal ir.ObjectLiteral) map[string]string {
	if literal.Spread == nil {
		return nil
	}
	getters := l.spreadGetters(literal)
	if len(getters) == 0 {
		return nil
	}
	left := map[string]string{}
	for _, class := range l.result.Classes {
		if class.Literal {
			continue
		}
		for _, accessor := range class.Accessors {
			if accessor.Getter < 0 || !slices.Contains(getters, accessor.Getter) {
				continue
			}
			if _, known := left[accessor.Name]; !known {
				left[accessor.Name] = class.Name
			}
		}
	}
	return left
}

// spreadCopyRead stops read, a read of name, when its receiver may be a spread's copy left without name
// (spreadCopyGetters), or when the graph can't close its reach while some spread may leave name out of
// its copy (copied).
func (l *lowering) spreadCopyRead(reach *accessorReach, copied map[string]string, read ir.Property, name string) error {
	where := read.ViewWhere
	if where == "" {
		where = l.result.Source
	}
	allocations := reach.graph.ReachingAllocations(read.Object)
	if allocations.Unknown {
		if class, ok := copied[name]; ok {
			return &NotYet{Where: where, What: "a read of " + name + " whose receiver may be an object a spread copied from a " + class + " instance, whose " + name + " is a getter on its prototype: a spread copies own properties only, so the copy has no " + name + " though its type says it does"}
		}
		return nil
	}
	records := reach.graph.projectionIndex().records
	for _, site := range allocations.Sites {
		literal, ok := records[site]
		if !ok || literal.Spread == nil {
			continue
		}
		if class, left := l.spreadCopyGetters(literal)[name]; left {
			return &NotYet{Where: where, What: "a read of " + name + " from " + literal.SpreadReadiness + ", a spread's copy of a " + class + " instance, whose " + name + " is a getter on its prototype: a spread copies own properties only, so the copy has no " + name + " though its type says it does"}
		}
		if class, left := l.spreadWithSourceGetter(reach, literal, name, map[int]bool{}); left {
			return &NotYet{Where: where, What: "a read of " + name + " from " + literal.SpreadReadiness + ", a spread's copy of a spread that copied a " + class + " instance, whose " + name + " is a getter on its prototype: a spread copies own properties only, so the copy has no " + name + " though its type says it does"}
		}
	}
	return nil
}

// spreadWithSourceGetter is the class whose name getter a SpreadWith literal reaching literal's source may leave
// out of its copy, through any spread between them. A copy is judged by the type its own spread sees, and a
// SpreadWith literal's type holds the names it adds, which the class's instances don't have, so { ...inner }
// over inner = { ...built, extra: 1 } can't find built's getter by its own type: the getter is the SpreadWith
// literal's to say (spreadCopyGetters). A reach the graph can't close fails closed on the getters some SpreadWith
// literal in the program may leave out (spreadWithCopied), so a program without one stops nothing new here.
func (l *lowering) spreadWithSourceGetter(reach *accessorReach, literal ir.ObjectLiteral, name string, seen map[int]bool) (string, bool) {
	allocations := reach.graph.ReachingAllocations(literal.Spread)
	if allocations.Unknown {
		class, left := l.spreadWithCopied[name]
		return class, left
	}
	records := reach.graph.projectionIndex().records
	for _, site := range allocations.Sites {
		source, ok := records[site]
		if !ok || source.Spread == nil || seen[site] {
			continue
		}
		seen[site] = true
		if class, left := l.spreadCopyGetters(source)[name]; left && source.SpreadWith {
			return class, true
		}
		if class, left := l.spreadWithSourceGetter(reach, source, name, seen); left {
			return class, true
		}
	}
	return "", false
}
