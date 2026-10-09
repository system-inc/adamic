package ir

import "reflect"

// ObjectSpreadMayThrow follows source storage, rather than making every spread
// exceptional because an unrelated getter exists. Unknown producers conservatively
// include the program's enumerable literal accessors. Class prototype getters are
// not own enumerable properties.
func (p *Program) ObjectSpreadMayThrow(literal ObjectLiteral) bool {
	if literal.Spread == nil {
		return false
	}
	throwing := false
	for _, class := range p.Classes {
		for _, accessor := range class.Accessors {
			throwing = throwing || class.Literal && accessor.Getter >= 0 && p.Functions[accessor.Getter].MayThrow
		}
	}
	if !throwing {
		return false
	}
	var data func(Expression, map[int]bool) bool
	data = func(value Expression, seen map[int]bool) bool {
		switch value := value.(type) {
		case ObjectLiteral:
			return value.Class == 0
		case Defined:
			return data(value.Value, seen)
		case Narrow:
			return data(value.Value, seen)
		case Unwrap:
			return data(value.Value, seen)
		case Conditional:
			return data(value.WhenTrue, seen) && data(value.WhenNot, seen)
		case Read:
			if p.Locals[value.Local].ExpressionAssigned {
				return false
			}
			for _, function := range p.Functions {
				for _, parameter := range function.Parameters {
					if parameter == value.Local {
						return false
					}
				}
			}
			if seen[value.Local] {
				return false
			}
			seen[value.Local] = true
			defer delete(seen, value.Local)
			found, safe := false, true
			visit := func(node any) {
				var stored Expression
				switch node := node.(type) {
				case Declare:
					if node.Local == value.Local && !node.Uninitialized {
						stored = node.Value
					}
				case Assign:
					if node.Local == value.Local {
						stored = node.Value
					}
				}
				if stored != nil {
					found = true
					safe = safe && data(stored, seen)
				}
			}
			spreadStorageWalk(reflect.ValueOf(p.Main), visit)
			for _, function := range p.Functions {
				spreadStorageWalk(reflect.ValueOf(function.Body), visit)
			}
			return found && safe
		}
		return false
	}
	return !data(literal.Spread, map[int]bool{})
}

func spreadStorageWalk(value reflect.Value, visit func(any)) {
	switch value.Kind() {
	case reflect.Interface:
		if !value.IsNil() {
			spreadStorageWalk(value.Elem(), visit)
		}
	case reflect.Struct:
		visit(value.Interface())
		for index := 0; index < value.NumField(); index++ {
			spreadStorageWalk(value.Field(index), visit)
		}
	case reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			spreadStorageWalk(value.Index(index), visit)
		}
	}
}
