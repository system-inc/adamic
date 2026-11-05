package javascript

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

const viewArraysRuntime = `
const adamicArrayStorage = new WeakMap();
const adamicViewArray = (array, references, kind) => { adamicArrayStorage.set(array,{references,kind}); return array; };
// Checked per read, never cached: an alias can push between reads.
const adamicViewElement = (array,index,relative,check) => {
 if(relative){index=Math.trunc(index)||0;if(index<0)index+=array.length;}
 if(!Number.isInteger(index)||index<0||index>=array.length)return undefined;
 const contract=adamicArrayContracts?.[check.Contract-1];
 const storage=adamicArrayStorage.get(array);
 const references=contract && [3,4,5,6,8,10].includes(contract.Of);
 const fail=kind=>panic('element read failed: '+check.Where+' is not a '+check.ViewType+'; found '+kind);
 if(check.Contract && storage && storage.references!==references)fail(storage.kind);
 const value=array[index];
 if(check.Contract && !adamicUntaggedPlainSelect(value,adamicArrayContracts,check.Contract,check.Where,check.ViewType,true))fail(value===null?'null':Array.isArray(value)?'array':value instanceof Map?'Map':adamicTypeOf(value));
 return value;
};
const adamicCheckedArrayRoad = (array,callback,initial,method,check,admitsUndefined,elementType) => {
 const count=array.length, mapped=[];
 let carried=initial;
 const last=method==='findLast'||method==='findLastIndex';
 const search=last||method==='find'||method==='findIndex';
 for(let step=0;step<count;step++){
  const index=last?count-1-step:step;
  if(index>=array.length){
   if(method==='map')panic('map: the array shrank while it was being mapped');
   if(!search)continue;
   if(!admitsUndefined)panic(method+': index '+index+' is undefined; element type '+elementType+' does not admit undefined');
  }
  const element=adamicViewElement(array,index,false,check);
  const answer=adamicCall(callback,method==='reduce'?[carried,element,index,array]:[element,index,array]);
  if(method==='reduce')carried=answer;
  else if(method==='map')mapped.push(answer);
  else if(method==='filter'){if(answer)mapped.push(element);}
  else if(method==='some'&&answer)return true;
  else if(method==='every'&&!answer)return false;
  else if(search&&answer){
   if(method==='find'||method==='findLast')return adamicViewElement(adamicViewArray([element],referencesForCheck(check),'reference storage'),0,false,check);
   return index;
  }
 }
 return method==='map'||method==='filter'?mapped:method==='reduce'?carried:method==='some'?false:method==='every'?true:search?(method==='find'||method==='findLast'?undefined:-1):undefined;
};
const referencesForCheck=check=>[3,4,5,6,8,10].includes(adamicArrayContracts?.[check.Contract-1]?.Of || 10);
`

func (e *emitter) viewArrayRuntime() string {
	contracts, err := json.Marshal(e.program.ViewContracts)
	if err != nil {
		panic(err)
	}
	return unionCallableShapeRuntime + viewUntaggedPlainRuntime + "\nconst adamicUntaggedCallableRecorded=(value)=>" + e.unionCallableRecorded("value") + ";\nconst adamicArrayContracts=" + string(contracts) + ";\n" + viewArraysRuntime
}

func (e *emitter) checkedElement(check *ir.ViewElementCheck, array, index string, relative bool) string {
	encoded, err := json.Marshal(check)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("adamicViewElement(%s,%s,%t,%s)", array, index, relative, encoded)
}

func (e *emitter) checkedArrayRoad(array, callback, initial ir.Expression, method string, check *ir.ViewElementCheck, admits bool, elementType string) string {
	encoded, err := json.Marshal(check)
	if err != nil {
		panic(err)
	}
	start := "undefined"
	if initial != nil {
		start = e.value(initial)
	}
	return fmt.Sprintf("adamicCheckedArrayRoad(%s,%s,%s,%s,%s,%t,%s)", e.value(array), e.value(callback), start, quote(method), encoded, admits, quote(elementType))
}

// Storage metadata is private and preserves the real array's identity.
func (e *emitter) value(expression ir.Expression) string {
	value := e.valueRaw(expression)
	var element ir.Type
	switch node := expression.(type) {
	case ir.ArrayLiteral:
		element = node.Element
	case ir.ArrayMap:
		element = node.Result
	case ir.ArrayVisit:
		if node.Method == "filter" {
			element = node.Element
		}
	case ir.ArraySlice:
		element = node.Element
	case ir.ArrayConcat:
		element = node.Element
	case ir.ArrayFrom:
		element = node.Element
	case ir.ArrayFill:
		element = node.Element
	case ir.ArraySplice:
		element = node.Element
	case ir.CodePoints, ir.ProgramArguments:
		element = ir.String
	}
	if element == 0 {
		return value
	}
	kind := "reference storage"
	if !element.IsReference() {
		kind = "number"
		if element == ir.Boolean {
			kind = "boolean"
		}
	}
	return fmt.Sprintf("adamicViewArray(%s,%t,%s)", value, element.IsReference(), quote(kind))
}
