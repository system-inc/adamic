// Write only the declared schema, looking up slots by name in the runtime layout.
#include "json_encode.h"
#include <stdlib.h>
#include <string.h>

typedef struct encode_builder {
 char *bytes;
 size_t length, capacity, units;
} encode_builder;

static void append(encode_builder *w, const char *bytes, size_t length, size_t units) {
	adamic_string_check_length((double)w->units + (double)units);
	if (length > SIZE_MAX - w->length) {
		static const char message[] = "RangeError: Invalid string length";
		adamic_panic(message, sizeof message - 1);
	}
	size_t needed = w->length + length;
	if (needed > w->capacity) {
		size_t capacity = w->capacity;
		while (capacity < needed) {
			capacity = capacity > SIZE_MAX / 2 ? needed : capacity * 2;
		}
		char *grown = realloc(w->bytes, capacity);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		w->bytes = grown;
		w->capacity = capacity;
	}
	if (length != 0) { memcpy(w->bytes + w->length, bytes, length); }
	w->length += length;
	w->units += units;
}
static void ascii(encode_builder *w, const char *text) {
	size_t length = strlen(text);
	append(w, text, length, length);
}
static void code_point(encode_builder *w, unsigned code) {
	char bytes[4];
	size_t count;
	if (code < 0x80) { bytes[0] = (char)code; count = 1; }
	else if (code < 0x800) {
		bytes[0] = (char)(0xc0 | (code >> 6)); bytes[1] = (char)(0x80 | (code & 63)); count = 2;
	} else if (code < 0x10000) {
		bytes[0] = (char)(0xe0 | (code >> 12)); bytes[1] = (char)(0x80 | ((code >> 6) & 63)); bytes[2] = (char)(0x80 | (code & 63)); count = 3;
	} else {
		bytes[0] = (char)(0xf0 | (code >> 18)); bytes[1] = (char)(0x80 | ((code >> 12) & 63));
		bytes[2] = (char)(0x80 | ((code >> 6) & 63)); bytes[3] = (char)(0x80 | (code & 63)); count = 4;
	}
	append(w, bytes, count, code >= 0x10000 ? 2 : 1);
}
static void quote(encode_builder *w, const adamic_string *text) {
	static const char hex[] = "0123456789abcdef";
	ascii(w, "\"");
	size_t length = (size_t)adamic_string_length(text);
	for (size_t index = 0; index < length; index++) {
		unsigned code = (unsigned)adamic_string_char_code_at(text, (double)index);
		switch (code) {
		case '"': ascii(w, "\\\""); continue;
		case '\\': ascii(w, "\\\\"); continue;
		case '\b': ascii(w, "\\b"); continue;
		case '\f': ascii(w, "\\f"); continue;
		case '\n': ascii(w, "\\n"); continue;
		case '\r': ascii(w, "\\r"); continue;
		case '\t': ascii(w, "\\t"); continue;
		default: break;
		}
		if (code >= 0xd800 && code <= 0xdbff && index + 1 < length) {
			unsigned low = (unsigned)adamic_string_char_code_at(text, (double)(index + 1));
			if (low >= 0xdc00 && low <= 0xdfff) {
				code_point(w, 0x10000 + ((code - 0xd800) << 10) + low - 0xdc00);
				index++;
				continue;
			}
		}
		if (code < 0x20 || (code >= 0xd800 && code <= 0xdfff)) {
			char escape[] = {'\\', 'u', hex[(code >> 12) & 15], hex[(code >> 8) & 15], hex[(code >> 4) & 15], hex[code & 15]};
			append(w, escape, sizeof escape, sizeof escape);
		} else { code_point(w, code); }
	}
	ascii(w, "\"");
}

// A slot's storage may be an optional scalar, even though its schema names the
// present type. Layout metadata distinguishes scalar slots from undefined NULL.
static bool encode_field(const adamic_object *object, const adamic_decode_field *field,
 const adamic_decode_node *type, adamic_value *value) {
 size_t position=object->shape->count;
 for (size_t i=0;i<object->shape->count;i++) {
  if (strcmp(object->shape->names[i],field->name)==0) { position=i; break; }
 }
 if (position==object->shape->count) { return false; }
 *value=object->slots[position];
 if (!field->optional) { return true; }
 if (object->shape->references[position]) {
  if (value->reference==NULL) { return false; }
  if (type->representation==1) { value->number=((adamic_number_box *)value->reference)->number; }
  else if (type->representation==2) { value->boolean=((adamic_boolean_box *)value->reference)->boolean; }
 } else if (type->representation==1) {
  adamic_maybe_number number=adamic_maybe_number_unpack(value->number);
  if (!number.present) { return false; }
  value->number=number.number;
 }
 return true;
}
static bool encode_literal(adamic_value value,int representation,const adamic_decode_node *type) {
 if (representation!=type->representation) { return false; }
 switch (representation) {
 case 1: return value.number==type->number;
 case 2: return value.boolean==type->boolean;
 case 3: return adamic_string_equal(value.reference,type->literal);
 default: return false;
 }
}
static void encode_write(encode_builder *builder, adamic_value value,
 const adamic_decode_schema *schema,size_t index,size_t depth) {
 ADAMIC_CHECK_STACK();
 const adamic_decode_node *type=&schema->nodes[index];
 if (strcmp(type->kind,"union")==0) {
  int representation=type->representation;
  if (representation==10) {
   const adamic_heap *heap=value.reference;
   if (heap==NULL) { adamic_panic("encodeJson: undefined union",sizeof "encodeJson: undefined union"-1); }
   switch (heap->kind) {
   case adamic_kind_number: representation=1;value.number=((const adamic_number_box *)heap)->number;break;
   case adamic_kind_boolean: representation=2;value.boolean=((const adamic_boolean_box *)heap)->boolean;break;
   case adamic_kind_string: representation=3;break;
   case adamic_kind_object: representation=4;break;
   default: adamic_panic("encodeJson: invalid union representation",sizeof "encodeJson: invalid union representation"-1);
   }
  }
  for (size_t i=0;i<type->child_count;i++) {
   size_t child=type->children[i];const adamic_decode_node *member=&schema->nodes[child];
   bool match=strcmp(member->kind,"literal")==0?encode_literal(value,representation,member):representation==member->representation;
   if (match && strcmp(member->kind,"object")==0 && type->discriminant[0]!='\0') {
    match=false;
    for (size_t f=0;f<member->field_count;f++) {
     const adamic_decode_field *field=&member->fields[f];
     if (strcmp(field->name,type->discriminant)!=0) { continue; }
     const adamic_decode_node *literal=&schema->nodes[field->node];adamic_value tag={0};
     match=encode_field(value.reference,field,literal,&tag) && encode_literal(tag,literal->representation,literal);
     break;
    }
   }
   if (match) { encode_write(builder,value,schema,child,depth);return; }
  }
  adamic_panic("encodeJson: value does not match its declared union",sizeof "encodeJson: value does not match its declared union"-1);
 }
 if (strcmp(type->kind,"array")==0 || strcmp(type->kind,"tuple")==0 || strcmp(type->kind,"object")==0) {
  if (strcmp(type->kind,"array")==0) {
   const adamic_array *array=value.reference;
   ascii(builder,"[");
   for (size_t i=0;i<array->length;i++) {
    if (i!=0) { ascii(builder,","); }
    encode_write(builder,array->elements[i],schema,type->children[0],depth+1);
   }
   ascii(builder,"]");return;
  }
  bool tuple=strcmp(type->kind,"tuple")==0;
  const adamic_object *object=value.reference;
  ascii(builder,tuple?"[":"{");size_t written=0;
  for (size_t i=0;i<type->field_count;i++) {
   const adamic_decode_field *field=&type->fields[i];adamic_value child={0};
   if (!encode_field(object,field,&schema->nodes[field->node],&child)) {
    if (field->optional) { continue; }
    adamic_panic("encodeJson: required field missing",sizeof "encodeJson: required field missing"-1);
   }
   if (written++!=0) { ascii(builder,","); }
   if (!tuple) {
    adamic_string name={{0,adamic_kind_string,0},strlen(field->name),field->name,0,NULL,NULL,0};
    quote(builder,&name);ascii(builder,":");
   }
   encode_write(builder,child,schema,field->node,depth+1);
  }
  ascii(builder,tuple?"]":"}");return;
 }
 switch (type->representation) {
 case 1: {
  if (!isfinite(value.number)) { ascii(builder,"null");return; }
  char bytes[ADAMIC_NUMBER_FORMAT_MAX];size_t length=adamic_number_format(value.number,bytes);
  append(builder,bytes,length,length);return;
 }
 case 2: ascii(builder,value.boolean?"true":"false");return;
 case 3: quote(builder,value.reference);return;
 default: adamic_panic("encodeJson: invalid descriptor",sizeof "encodeJson: invalid descriptor"-1);
 }
}
adamic_string *adamic_json_encode(adamic_value value,const adamic_decode_schema *schema) {
 encode_builder builder={NULL,0,64,0};
 builder.bytes=malloc(builder.capacity);
 if (builder.bytes==NULL) { adamic_panic("out of memory",sizeof "out of memory"-1); }
 encode_write(&builder,value,schema,schema->root,0);
 adamic_string *result=adamic_string_allocate(builder.length);
 if (builder.length!=0) { memcpy((char *)result->bytes,builder.bytes,builder.length); }
 free(builder.bytes);return result;
}
