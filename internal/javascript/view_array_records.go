package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) emitViewArrayRecord(record ir.ArrayRecord) string {
	return fmt.Sprintf("((array, source) => { for (const name of Object.keys(source)) adamicWriteField(array, name, adamicReadField(source, name, %s)); adamicRecordFieldTypes(array, {...adamicFieldRepresentations.get(source)}); adamicRecordSlotContracts(array, {...adamicSlotContracts.get(source)}, 'array'); return array; })(%s, %s)", quote(record.Where), e.value(record.Array), e.value(record.Properties))
}

func (e *emitter) emitViewArrayProperties(properties ir.ArrayProperties) string {
	return e.value(properties.Array)
}
