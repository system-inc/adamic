"""Run semantic mutants; require a Node comparison failure and restore exact bytes."""
import os,subprocess,sys
from pathlib import Path
root=Path(__file__).resolve().parents[3]
logs=root/'docs/library-object/unit3/mutants';logs.mkdir(exist_ok=True)
cases=[
 ('plain-object-view-proof','internal/lower/prototype.go','if l.exactObject(receiver, 0) {','if true {','TestObjectPrototypeRepresentationsStayDistinct$'),
 ('numeric-key-sign','internal/lower/library_object_descriptors.go','ir.NumberToString{Value: ir.Read{Local: parameters[1], Of: ir.Number}}','ir.NumberToString{Value: ir.MathCall{Function: "abs", Arguments: []ir.Expression{ir.Read{Local: parameters[1], Of: ir.Number}}}}','TestNativeAgreesWithNode/internal/oracle/testdata/library_object_numeric_keys.a$'),
 ('numeric-key-typed-slot-proof','internal/lower/library_object_descriptors.go','len(l.checker.GetPropertiesOfType(targetType)) == 0','len(l.checker.GetPropertiesOfType(targetType)) >= 0','TestObjectNumericTypedSlotStaysRefused$'),
 ('library-typeerror-name','internal/native/runtime/object_descriptors.c','static adamic_string name = ADAMIC_STRING("TypeError");','static adamic_string name = ADAMIC_STRING("Error");','TestNativeAgreesWithNode/internal/oracle/testdata/library_object_descriptor_errors.a$'),
 ('own-keys-array-boundary','internal/native/runtime/object_names.c','((const adamic_array *)value)->length : 0','(((const adamic_array *)value)->length == 0 ? 0 : ((const adamic_array *)value)->length - 1) : 0','TestNativeAgreesWithNode/internal/oracle/testdata/library_object_own_keys_receivers.a$'),
 ('descriptor-same-value-zero','internal/native/runtime/object_descriptors.c','left == right && (left != 0 || signbit(left) == signbit(right))','left == right','TestNativeAgreesWithNode/internal/oracle/testdata/library_object_descriptor_same_value.a$'),
 ('descriptor-new-enumerable','internal/native/runtime/object_descriptors.c','bool old_enumerable = exists && (entry == NULL || entry->enumerable);','bool old_enumerable = !exists || (entry == NULL || entry->enumerable);','TestNativeAgreesWithNode/internal/oracle/testdata/library_object_descriptor_order.a$'),
 ('define-properties-index-order','internal/lower/library_object_descriptors.go','return xi && x < y','return xi && x > y','TestNativeAgreesWithNode/internal/oracle/testdata/library_object_define_properties.a$'),
 ('collection-extensibility','internal/native/runtime/object_integrity.c','if (value->kind == adamic_kind_map) return collection_level(value) == NULL;','if (value->kind == adamic_kind_map) return collection_level(value) != NULL;','TestNativeAgreesWithNode/internal/oracle/testdata/library_object_collection_integrity.a$'),
 ('has-own-hidden-key','internal/native/runtime/object.c','return adamic_object_descriptor_has(object, name);','return false;','TestNativeAgreesWithNode/internal/oracle/testdata/library_object_descriptor_errors.a$'),
 ('descriptor-consumer-proof','internal/lower/library_object_descriptors.go','if hazard := l.objectDescriptorHazard(); hazard != "" {','if hazard := l.objectDescriptorHazard(); hazard == "mutant" {','TestObjectDescriptorConsumerStaysRefused$'),
 ('descriptor-map-proof','internal/lower/library_object_descriptors.go','if literal == nil || l.objectDescriptorWasMutated(literal) {','if literal == nil {','TestObjectDescriptorMapStaysRefused$'),
]
if len(sys.argv)>1: cases=[case for case in cases if case[0] in sys.argv[1:]]
for name,relative,old,new,pattern in cases:
 path=root/relative;original=path.read_bytes();assert old in original.decode(),name
 try:
  path.write_text(original.decode().replace(old,new,1))
  with (logs/(name+'.log')).open('w') as out:
   result=subprocess.run(['go','test','-count=1','-timeout','5m','./internal/oracle','-run',pattern],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=out,stderr=subprocess.STDOUT)
  output=(logs/(name+'.log')).read_text()
  caught=result.returncode!=0 and ('stdout differs' in output or 'exit codes differ' in output) and all(x not in output for x in ['Lower:','Load:','AddressSanitizer','runtime error:','error:'])
  print(name+': '+('caught by Node comparison' if caught else 'NOT PROVEN'),flush=True)
  if not caught:raise SystemExit(output)
 finally:path.write_bytes(original)
