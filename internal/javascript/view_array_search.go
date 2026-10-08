package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) emitViewArraySearch(search ir.ArraySearch) string {
	from := "0"
	if search.From != nil {
		from = e.value(search.From)
	}
	start := "Math.trunc(from) || 0"
	relative := "if (start < 0) start += count;"
	bound := "Math.max(start, 0)"
	condition, update := "index < count", "index++"
	found, missing := "index", "-1"
	if search.Last {
		bound = "Math.min(start, count - 1)"
		condition, update = "index >= 0", "index--"
		if search.From == nil {
			start, relative = "count - 1", ""
		}
	}
	if search.Includes {
		found, missing = "true", "false"
	}
	return fmt.Sprintf(`((array, wanted, from, check) => {
    const count = array.length;
    let start = %s;
    %s
    start = %s;
    for (let index = start; %s; %s) {
        if (!%t && !(index in array)) continue;
        const value = index in array ? check(array[index]) : undefined;
        if (value === wanted || (%t && Number.isNaN(value) && Number.isNaN(wanted))) return %s;
    }
    return %s;
})(%s, %s, %s, %s)`, start, relative, bound, condition, update, search.Includes, search.Includes, found, missing, e.value(search.Array), e.value(search.Value), from, e.viewArrayChecker(search.ViewRead))
}
