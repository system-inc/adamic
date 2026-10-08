package javascript

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

func (e *emitter) mapNominalGraphCertificate(id ir.ViewContractID, value, where string) string {
	type field struct {
		Name     string `json:"name"`
		Schema   int    `json:"schema"`
		Optional bool   `json:"optional"`
	}
	type schema struct {
		Name      string  `json:"name"`
		Class     int     `json:"class"`
		Null      bool    `json:"null"`
		Undefined bool    `json:"undefined"`
		Fields    []field `json:"fields"`
	}
	rows := make([]schema, len(e.program.ViewContracts))
	for index, contract := range e.program.ViewContracts {
		if !ir.HasMapNominalWitness(e.program, ir.ViewContractID(index+1)) {
			continue
		}
		nullable, undefined := contract.Null, contract.Undefined
		if contract.Kind == ir.ViewNullable {
			contract = e.program.ViewContracts[contract.Element-1]
		}
		row := schema{Name: contract.Name, Class: contract.NominalClass, Null: nullable, Undefined: undefined || contract.Undefined, Fields: []field{}}
		for _, child := range contract.Fields {
			if ir.HasMapNominalWitness(e.program, child.Contract) {
				row.Fields = append(row.Fields, field{child.Name, int(child.Contract) - 1, child.Optional})
			}
		}
		rows[index] = row
	}
	data, err := json.Marshal(rows)
	if err != nil {
		panic(err)
	}
	return "((root)=>{const schemas=" + string(data) + ",work=[[root," + strconv.Itoa(int(id)-1) + "]],seen=new WeakMap();while(work.length){const [entry,id]=work.pop(),schema=schemas[id];const fail=()=>panic(" + quote("Map nominal producer failed: "+where+" expected ") + "+schema.name+', found value without its class identity');if(entry===null||entry===undefined){if(entry===null?!schema.null:!schema.undefined)fail();continue;}if(typeof entry!=='object'||Array.isArray(entry)||entry instanceof Map||entry instanceof Set||(schema.class&&!adamicInstanceOf(entry,schema.class)))fail();let obligations=seen.get(entry);if(!obligations){obligations=new Set();seen.set(entry,obligations);}if(obligations.has(id))continue;obligations.add(id);for(const field of schema.fields){if(!(field.name in entry)){if(field.optional)continue;panic('field read failed: '+field.name+' expected '+schemas[field.schema].name+', found absent');}work.push([entry[field.name],field.schema]);}}return root;})(" + value + ")"
}
