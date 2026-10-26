package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// constantNumericArray stores numeric literal data outside the initializer's executable
// code. Each evaluation still makes an ordinary, independently mutable array with
// the same capacity and ownership. Reads, calls, spreads and conversions stay on
// the normal evaluation path; no runtime expression is treated as a constant.
// Functions keep literals smaller than constantNumbers inline, avoiding a call and
// copy loop on hot paths. Module initialization can use tables from two elements;
// single-element literals keep edits such as [14].length in their source unit.
func (e *emitter) constantNumericArray(literal ir.ArrayLiteral) (string, bool) {
	if literal.Element != ir.Number || literal.Spread != nil || len(literal.Elements) < 2 {
		return "", false
	}
	if e.function != nil && len(literal.Elements) < constantNumbers {
		return "", false
	}
	values := make([]string, len(literal.Elements))
	for index, element := range literal.Elements {
		value, ok := constantTableNumber(element)
		if !ok {
			return "", false
		}
		values[index] = cNumber(value)
	}
	name := e.numberTable(values)
	helper := e.numberTableCopy()
	array := e.own(ir.Array, fmt.Sprintf("%s(%d, %s)", helper, len(values), name))
	e.adoptGraph(array, "sizeof *"+array, e.graphTypes(literal.GraphTypes) && !literal.ProgramRegion)
	e.programArray(array, literal.ProgramRegion)
	return array, true
}

func constantTableNumber(expression ir.Expression) (float64, bool) {
	switch expression := expression.(type) {
	case ir.NumberConstant:
		return expression.Value, true
	case ir.Unary:
		if expression.Operator == ir.Negate || expression.Operator == ir.Plus {
			if value, ok := constantTableNumber(expression.Operand); ok {
				if expression.Operator == ir.Negate {
					value = -value
				}
				return value, true
			}
		}
	}
	return 0, false
}

func (e *emitter) numberTableCopy() string {
	helper, found := e.sharedDeclaration("adamic_initialize_number_table_copy", "numbers")
	if !found {
		e.declarations = append(e.declarations, fmt.Sprintf(`static adamic_array *%s(size_t count, const double *values) __attribute__((noinline));
static adamic_array *%s(size_t count, const double *values) {
 adamic_array *array = adamic_array_new(count, false);
 for (size_t index = 0; index < count; index++) {
  adamic_array_push(array, (adamic_value){.number = values[index]});
 }
 return array;
}`, helper, helper))
	}
	return helper
}

// constantRecords batches closed literal records, never imported reads or other
// runtime operands. Numeric array fields are reconstructed before their record;
// strings point to the existing immortal literals. Every record and nested array
// is newly allocated, with its usual shape, mutable slots and reference counts.
// Keep this top-level only: function literals may belong to reuse or region plans.
// Graph, Program, record and checked-field layouts keep ordinary emission so their
// allocation-site ownership and per-slot metadata remain under the existing planner.
func (e *emitter) constantRecords(literal ir.ArrayLiteral) (string, bool) {
	if e.function != nil || e.fieldTypesNeeded() || e.dynamicProperties() || literal.ProgramRegion || e.graphTypes(literal.GraphTypes) || literal.Element != ir.Object || literal.Spread != nil || len(literal.Elements) < 16 {
		return "", false
	}
	type run struct {
		begin, end int
		rows       []constantRecord
	}
	var runs []run
	for begin := 0; begin < len(literal.Elements); {
		first, ok := e.constantRecordOf(literal.Elements[begin])
		if !ok {
			begin++
			continue
		}
		rows := []constantRecord{first}
		end := begin + 1
		for end < len(literal.Elements) {
			row, ok := e.constantRecordOf(literal.Elements[end])
			if !ok || row.layout != first.layout {
				break
			}
			rows = append(rows, row)
			end++
		}
		if len(rows) >= 8 {
			runs = append(runs, run{begin, end, rows})
		}
		begin = end
	}
	if len(runs) == 0 {
		return "", false
	}
	array := e.own(ir.Array, fmt.Sprintf("adamic_array_new(%d, true)", len(literal.Elements)))
	next := 0
	for index := 0; index < len(literal.Elements); {
		if next < len(runs) && index == runs[next].begin {
			group := runs[next]
			e.emitConstantRecords(array, group.rows)
			index = group.end
			next++
		} else {
			value := e.value(literal.Elements[index])
			e.line("adamic_array_push(%s, (adamic_value){.reference = %s});", array, retained(value))
			index++
		}
	}
	return array, true
}

// Limit specialization so the generated constructor itself stays bounded. Larger
// shapes use the normal stores, which moduleMain outlines into statement chunks.
const constantRecordFields = 16

type constantRecord struct {
	literal ir.ObjectLiteral
	layout  string
	slots   []string
	arrays  map[int][]string
}

func (e *emitter) constantRecordOf(expression ir.Expression) (constantRecord, bool) {
	literal, ok := expression.(ir.ObjectLiteral)
	if !ok || literal.Record || literal.ProgramRegion || e.graphTypes(literal.GraphTypes) || literal.Class != 0 || literal.Spread != nil || len(literal.Methods) != 0 || literal.Tuple || len(literal.Fields) == 0 || len(literal.Fields) > constantRecordFields {
		return constantRecord{}, false
	}
	row := constantRecord{literal: literal, arrays: map[int][]string{}}
	for index, field := range literal.Fields {
		if field.Private || field.Uninitialized {
			return constantRecord{}, false
		}
		row.layout += fmt.Sprintf("%q/%d;", field.Name, field.Value.Type())
		if value, ok := constantTableNumber(field.Value); ok {
			row.slots = append(row.slots, "{.number = "+cNumber(value)+"}")
			continue
		}
		switch value := field.Value.(type) {
		case ir.BooleanConstant:
			row.slots = append(row.slots, fmt.Sprintf("{.boolean = %t}", value.Value))
		case ir.StringConstant:
			row.slots = append(row.slots, "{.reference = &"+stringName(e.program.Strings[value.Index])+"}")
		case ir.ArrayLiteral:
			if value.Element != ir.Number || value.Spread != nil || value.ProgramRegion || e.graphTypes(value.GraphTypes) {
				return constantRecord{}, false
			}
			values := []string{}
			for _, element := range value.Elements {
				number, ok := constantTableNumber(element)
				if !ok {
					return constantRecord{}, false
				}
				values = append(values, cNumber(number))
			}
			row.arrays[index] = values
			row.slots = append(row.slots, "{.reference = NULL}")
		default:
			return constantRecord{}, false
		}
	}
	return row, true
}

func (e *emitter) emitConstantRecords(array string, rows []constantRecord) {
	first := rows[0]
	shape := e.literalShape(first.literal)
	var slots, numbers, offsets []string
	offsets = append(offsets, "0")
	for _, row := range rows {
		slots = append(slots, "{"+strings.Join(row.slots, ", ")+"}")
		for index := range row.slots {
			if values, ok := row.arrays[index]; ok {
				numbers = append(numbers, values...)
				offsets = append(offsets, fmt.Sprint(len(numbers)))
			}
		}
	}
	data := strings.Join(slots, ",\n\t")
	table, found := e.sharedDeclaration("adamic_record_table", data)
	if !found {
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_value %s[%d][%d] = {\n\t%s\n};", table, len(rows), len(first.slots), data))
	}
	numbersName, offsetsName := "NULL", "NULL"
	if len(first.arrays) > 0 {
		// The empty sentinel is never read, but C11 requires nonzero array size.
		if len(numbers) == 0 {
			numbers = append(numbers, cNumber(0))
		}
		numbersName = e.numberTable(numbers)
		offsetsName = e.lengthTable(offsets)
	}
	helper, found := e.sharedDeclaration("adamic_initialize_record_table_copy", first.layout)
	if !found {
		lines := []string{fmt.Sprintf("static void %s(adamic_array *array, size_t count, const adamic_value values[][%d], const double *numbers, const size_t *offsets) {", helper, len(first.slots)),
			" (void)numbers; (void)offsets;", " for (size_t row = 0; row < count; row++) {"}
		part := 0
		for index := range first.slots {
			if _, ok := first.arrays[index]; ok {
				lines = append(lines, fmt.Sprintf("  size_t begin_%d = offsets[row * %d + %d], end_%d = offsets[row * %d + %d];", index, len(first.arrays), part, index, len(first.arrays), part+1),
					fmt.Sprintf("  adamic_array *field_%d = %s(end_%d - begin_%d, numbers + begin_%d);", index, e.numberTableCopy(), index, index, index))
				part++
			}
		}
		lines = append(lines, fmt.Sprintf("  adamic_object *object = adamic_object_new(&%s);", shape))
		for index, field := range first.literal.Fields {
			if _, ok := first.arrays[index]; ok {
				lines = append(lines, fmt.Sprintf("  object->slots[%d].reference = field_%d;", index, index))
			} else {
				lines = append(lines, fmt.Sprintf("  object->slots[%d] = values[row][%d];", index, index))
				if field.Value.Type().IsReference() {
					lines = append(lines, fmt.Sprintf("  adamic_retain(object->slots[%d].reference);", index))
				}
			}
		}
		lines = append(lines, "  adamic_array_push(array, (adamic_value){.reference = object});", " }", "}")
		prototype := strings.TrimSuffix(lines[0], " {") + " __attribute__((noinline));"
		e.declarations = append(e.declarations, prototype+"\n"+strings.Join(lines, "\n"))
	}
	e.line("%s(%s, %d, %s, %s, %s);", helper, array, len(rows), table, numbersName, offsetsName)
}
