// JSON grammar and conversion are independent of the Node reference. Scratch nodes own their
// strings and children; validation owns its partially built output until it succeeds.
#include "json_decode.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static adamic_string *new_text(const char *bytes, size_t size) {
 adamic_string *text=adamic_string_allocate(size);
 if (size!=0) { memcpy((char *)text->bytes,bytes,size); }
 return text;
}
static adamic_string *join(const adamic_string *a,const adamic_string *b) {
 adamic_string *parts[]={(adamic_string *)a,(adamic_string *)b};
 return adamic_string_concat(2,parts);
}
typedef enum { json_null, json_number, json_boolean, json_string, json_array, json_object } json_kind;
typedef struct json_node json_node;
struct json_node {
 json_kind kind;
 double number;
 bool boolean;
 adamic_string *string;
 size_t count;
 json_node **children;
 adamic_string **keys;
};
typedef struct json_reader {
 const adamic_string *text;
 size_t position;
 size_t length;
 const char *error;
} json_reader;
static void *memory(size_t size) {
 void *p = calloc(1, size == 0 ? 1 : size);
 if (p == NULL) { static const char m[]="out of memory"; adamic_panic(m,sizeof m-1); }
 return p;
}
static void dispose(json_node *n) {
 if (n == NULL) { return; }
 adamic_release(n->string);
 for (size_t i=0; i<n->count; i++) {
  dispose(n->children[i]);
  if (n->keys != NULL) { adamic_release(n->keys[i]); }
 }
 free(n->children); free(n->keys); free(n);
}
static unsigned peek(json_reader *r) {
 return r->position < r->length ? (unsigned)adamic_string_char_code_at(r->text,(double)r->position) : 0x10000;
}
static void whitespace(json_reader *r) {
 while (peek(r)==' ' || peek(r)=='\t' || peek(r)=='\n' || peek(r)=='\r') { r->position++; }
}
static bool failure(json_reader *r,const char *message) {
 if (r->error==NULL) { r->error=message; }
 return false;
}
static int hex(unsigned c) {
 if (c>='0' && c<='9') { return (int)(c-'0'); }
 if (c>='a' && c<='f') { return (int)(c-'a'+10); }
 if (c>='A' && c<='F') { return (int)(c-'A'+10); }
 return -1;
}
static adamic_string *read_string(json_reader *r) {
 r->position++;
 // At most one output unit per remaining input unit. fromCharCode preserves lone surrogates.
 double *units=memory((r->length-r->position+1)*sizeof *units);
 size_t count=0;
 while (r->position<r->length) {
  unsigned c=peek(r); r->position++;
  if (c=='"') { adamic_string *s=adamic_string_from_char_codes(count,units); free(units); return s; }
  if (c<0x20) { r->position--; failure(r,"control character in string"); break; }
  if (c=='\\') {
   if (r->position==r->length) { failure(r,"unterminated string"); break; }
   c=peek(r); r->position++;
   switch (c) {
   case '"': case '\\': case '/': break;
   case 'b': c=8; break; case 'f': c=12; break; case 'n': c=10; break;
   case 'r': c=13; break; case 't': c=9; break;
   case 'u': {
    c=0;
    for (size_t i=0; i<4; i++) {
     int digit=hex(peek(r));
     if (digit<0) { failure(r,"expected four hexadecimal digits"); break; }
     c=c*16+(unsigned)digit; r->position++;
    }
    break;
   }
   default: r->position--; failure(r,"invalid string escape"); break;
   }
  }
  if (r->error!=NULL) { break; }
  units[count++]=(double)c;
 }
 free(units); failure(r,"unterminated string"); return NULL;
}
static bool digit(unsigned c) { return c>='0' && c<='9'; }
static json_node *read_value(json_reader *r,size_t depth);
static void add(json_node *parent,json_node *child,adamic_string *key) {
 size_t count=parent->count;
 json_node **children=memory((count+1)*sizeof *children);
 if (count>0) { memcpy(children,parent->children,count*sizeof *children); }
 free(parent->children); parent->children=children; children[count]=child;
 if (parent->kind==json_object) {
  adamic_string **keys=memory((count+1)*sizeof *keys);
  if (count>0) { memcpy(keys,parent->keys,count*sizeof *keys); }
  free(parent->keys); parent->keys=keys; keys[count]=key;
 }
 parent->count++;
}
static json_node *container(json_reader *r,size_t depth,bool object) {
 if (depth>=ADAMIC_JSON_DEPTH) { failure(r,"nesting depth exceeds 128"); return NULL; }
 r->position++;
 json_node *n=memory(sizeof *n); n->kind=object?json_object:json_array;
 whitespace(r);
 unsigned closing=object?'}':']';
 if (peek(r)==closing) { r->position++; return n; }
 for (;;) {
  adamic_string *key=NULL;
  if (object) {
   if (peek(r)!='"') { failure(r,"expected object key"); break; }
   key=read_string(r); if (key==NULL) { break; }
   whitespace(r);
   if (peek(r)!=':') { adamic_release(key); failure(r,"expected ':'"); break; }
   r->position++; whitespace(r);
  }
  json_node *child=read_value(r,depth+1);
  if (child==NULL) { adamic_release(key); break; }
  // Last duplicate wins. Keeping scratch duplicates until disposal is harmless: lookup scans back.
  add(n,child,key); whitespace(r);
  if (peek(r)==closing) { r->position++; return n; }
  if (peek(r)!=',') { failure(r,object?"expected ',' or '}'":"expected ',' or ']'"); break; }
  r->position++; whitespace(r);
 }
 dispose(n); return NULL;
}
static json_node *read_value(json_reader *r,size_t depth) {
 whitespace(r);
 unsigned c=peek(r);
 if (c=='[' || c=='{') { return container(r,depth,c=='{'); }
 json_node *n=memory(sizeof *n);
 if (c=='"') { n->kind=json_string; n->string=read_string(r); if (n->string!=NULL) { return n; } }
 else if (c=='-' || digit(c)) {
  size_t start=r->position; n->kind=json_number;
  if (c=='-') { r->position++; }
  if (peek(r)=='0') { r->position++; }
  else if (peek(r)>='1' && peek(r)<='9') { while (digit(peek(r))) { r->position++; } }
  else { failure(r,"expected digit"); }
  if (r->error==NULL && peek(r)=='.') {
   r->position++;
   if (!digit(peek(r))) { failure(r,"expected digit"); }
   while (digit(peek(r))) { r->position++; }
  }
  if (r->error==NULL && (peek(r)=='e' || peek(r)=='E')) {
   r->position++;
   if (peek(r)=='+' || peek(r)=='-') { r->position++; }
   if (!digit(peek(r))) { failure(r,"expected digit"); }
   while (digit(peek(r))) { r->position++; }
  }
  if (r->error==NULL) {
   size_t size=r->position-start; char *bytes=memory(size+1);
   for (size_t i=0; i<size; i++) { bytes[i]=(char)adamic_string_char_code_at(r->text,(double)(start+i)); }
   // strtod is the runtime's decimal conversion too. The oracle checks the exact double bits.
   n->number=strtod(bytes,NULL); free(bytes); return n;
  }
 } else {
  const char *word=c=='t'?"true":c=='f'?"false":c=='n'?"null":NULL;
  if (word!=NULL) {
   size_t i=0;
   while (word[i]!='\0' && peek(r)==(unsigned char)word[i]) { i++; r->position++; }
   if (word[i]=='\0') { n->kind=c=='n'?json_null:json_boolean; n->boolean=c=='t'; return n; }
   failure(r,"invalid keyword");
  } else { failure(r,"expected JSON value"); }
 }
 dispose(n); return NULL;
}
static adamic_string *text_of(const char *s) { return new_text(s,strlen(s)); }
static adamic_string *syntax_message(json_reader *r) {
 size_t line=1,column=1;
 for (size_t i=0;i<r->position;i++) {
  if ((unsigned)adamic_string_char_code_at(r->text,(double)i)=='\n') { line++; column=1; }
  else { column++; }
 }
 char prefix[128]; int length=snprintf(prefix,sizeof prefix,"invalid JSON at line %zu column %zu: ",line,column);
 adamic_string *a=new_text(prefix,(size_t)length), *b=text_of(r->error);
 adamic_string *message=join(a,b); adamic_release(a); adamic_release(b); return message;
}
static json_node *field(const json_node *n,const char *name) {
 adamic_string *key=text_of(name); json_node *result=NULL;
 for (size_t i=n->count;i>0;i--) { if (adamic_string_equal(n->keys[i-1],key)) { result=n->children[i-1]; break; } }
 adamic_release(key); return result;
}
static const char *found(json_kind k) {
 const char *names[]={"null","number","boolean","string","array","object"}; return names[k];
}
static adamic_string *path_field(const adamic_string *path,const char *name,bool index) {
 adamic_string *a=text_of(index?"[":"."), *b=text_of(name), *c=text_of(index?"]":"");
 adamic_string *p=join(path,a), *q=join(p,b), *result=join(q,c);
 adamic_release(a); adamic_release(b); adamic_release(c); adamic_release(p); adamic_release(q); return result;
}
static void mismatch(adamic_string **error,const adamic_string *path,const char *expected,json_kind kind) {
 if (*error!=NULL) { return; }
 adamic_string *a=text_of("at "), *b=text_of(": expected "), *c=text_of(expected), *d=text_of(", found "), *f=text_of(found(kind));
 adamic_string *p=join(a,path), *q=join(p,b), *s=join(q,c), *t=join(s,d);
 *error=join(t,f);
 adamic_release(a); adamic_release(b); adamic_release(c); adamic_release(d); adamic_release(f);
 adamic_release(p); adamic_release(q); adamic_release(s); adamic_release(t);
}
static void missing(adamic_string **error,const adamic_string *path,const char *name) {
 adamic_string *a=text_of("at "),*b=text_of(": missing field "),*c=text_of(name);
 adamic_string *p=join(a,path),*q=join(p,b); *error=join(q,c);
 adamic_release(a);adamic_release(b);adamic_release(c);adamic_release(p);adamic_release(q);
}
static bool literal(const json_node *value,const adamic_decode_node *type) {
 if (type->representation==1) { return value->kind==json_number && value->number==type->number; }
 if (type->representation==2) { return value->kind==json_boolean && value->boolean==type->boolean; }
 return value->kind==json_string && adamic_string_equal(value->string,type->literal);
}
static bool reference(int representation) { return representation!=1 && representation!=2 && representation!=7; }
// A decoded object's shape lives in the object's own allocation, after the slots. It has no
// separate owner and needs no new heap kind, and absent optional fields take no slot at all.
static adamic_object *make_object(size_t count) {
 size_t size=sizeof(adamic_object)+count*sizeof(adamic_value)+sizeof(adamic_shape)+count*sizeof(char *)+count*sizeof(bool);
 adamic_object *o=adamic_allocate(size,adamic_kind_object);
 memset((char *)o+sizeof(adamic_heap),0,size-sizeof(adamic_heap));
 adamic_shape *shape=(adamic_shape *)(o->slots+count);
 shape->count=count;
 shape->names=(const char *const *)(shape+1);
 shape->references=(const bool *)(shape->names+count);
 o->shape=shape; return o;
}
// Field-read caches keep shape identity across object lifetimes. Canonical shapes therefore live
// until program exit; a shape embedded in a freed object could be mistaken for a later layout.
typedef struct decode_shape_entry {
 struct decode_shape_entry *next;
 adamic_shape shape;
} decode_shape_entry;
static decode_shape_entry *decode_shapes;
static adamic_shape *canonical_shape(const adamic_shape *shape) {
 for (decode_shape_entry *entry=decode_shapes;entry!=NULL;entry=entry->next) {
  if (entry->shape.count!=shape->count) { continue; }
  bool same=true;
  for (size_t i=0;i<shape->count;i++) {
   if (entry->shape.references[i]!=shape->references[i] || strcmp(entry->shape.names[i],shape->names[i])!=0) { same=false; break; }
  }
  if (same) { return &entry->shape; }
 }
 decode_shape_entry *entry=memory(sizeof *entry+shape->count*(sizeof(char *)+sizeof(bool)));
 entry->next=decode_shapes;entry->shape=*shape;
 entry->shape.names=(const char *const *)(entry+1);
 entry->shape.references=(const bool *)(entry->shape.names+shape->count);
 if (shape->count!=0) {
  memcpy((void *)entry->shape.names,shape->names,shape->count*sizeof(char *));
  memcpy((void *)entry->shape.references,shape->references,shape->count*sizeof(bool));
 }
 decode_shapes=entry;return &entry->shape;
}
__attribute__((destructor)) static void free_decode_shapes(void) {
 while (decode_shapes!=NULL) { decode_shape_entry *entry=decode_shapes;decode_shapes=entry->next;free(entry); }
}
static bool validate(const adamic_decode_schema *schema,size_t index,const json_node *value,const adamic_string *path,adamic_value *out,adamic_string **error) {
 const adamic_decode_node *type=&schema->nodes[index]; const char *kind=type->kind;
 if (strcmp(kind,"union")==0) {
  size_t selected=SIZE_MAX;
  json_node *tag=NULL;
  if (type->discriminant[0]!='\0' && value->kind==json_object) {
   tag=field(value,type->discriminant);
   if (tag==NULL) { missing(error,path,type->discriminant); return false; }
  }
  for (size_t i=0;i<type->child_count;i++) {
   size_t child=type->children[i]; const adamic_decode_node *member=&schema->nodes[child];
   bool match=false;
   if (strcmp(member->kind,"null")==0) { match=value->kind==json_null; }
   else if (strcmp(member->kind,"object")==0 && value->kind==json_object) {
    match=type->discriminant[0]=='\0';
    for (size_t f=0;tag!=NULL && f<member->field_count;f++) {
     if (strcmp(member->fields[f].name,type->discriminant)==0) { match=literal(tag,&schema->nodes[member->fields[f].node]); break; }
    }
   } else if (strcmp(member->kind,"literal")==0) { match=literal(value,member); }
   else { match=strcmp(member->kind,found(value->kind))==0; }
   if (match) { selected=child; break; }
  }
  if (selected==SIZE_MAX) {
   adamic_string *where=tag!=NULL?path_field(path,type->discriminant,false):adamic_retain((void *)path);
   mismatch(error,where,type->expected,tag!=NULL?tag->kind:value->kind); adamic_release(where); return false;
  }
  adamic_value result={0};
  if (!validate(schema,selected,value,path,&result,error)) { return false; }
  int of=schema->nodes[selected].representation;
  if (type->representation==10 && of==1) { result.reference=adamic_box_number(result.number); }
  else if (type->representation==10 && of==2) { result.reference=result.boolean?(void *)&adamic_box_true:(void *)&adamic_box_false; }
  *out=result; return true;
 }
 bool match=strcmp(kind,"literal")==0?literal(value,type):strcmp(kind,"tuple")==0?value->kind==json_array:strcmp(kind,found(value->kind))==0;
 if (!match) { mismatch(error,path,type->expected,value->kind); return false; }
 if (value->kind==json_null) { out->reference=NULL; return true; }
 if (value->kind==json_number) { out->number=value->number; return true; }
 if (value->kind==json_boolean) { out->boolean=value->boolean; return true; }
 if (value->kind==json_string) { out->reference=adamic_retain(value->string); return true; }
 if (strcmp(kind,"array")==0) {
  int of=schema->nodes[type->children[0]].representation;
  adamic_array *a=adamic_array_new(value->count,reference(of));
  for (size_t i=0;i<value->count;i++) {
   char name[32]; (void)snprintf(name,sizeof name,"%zu",i);
   adamic_string *where=path_field(path,name,true); adamic_value child={0};
   bool ok=validate(schema,type->children[0],value->children[i],where,&child,error); adamic_release(where);
   if (!ok) { adamic_release(a); return false; }
   adamic_array_push(a,child);
  }
  out->reference=a; return true;
 }
 bool tuple=strcmp(kind,"tuple")==0;
 if (tuple && value->count!=type->field_count) { mismatch(error,path,type->expected,value->kind); return false; }
 size_t count=0;
 for (size_t i=0;i<type->field_count;i++) { if (tuple || field(value,type->fields[i].name)!=NULL) { count++; } }
 adamic_object *o=make_object(count); size_t written=0;
 for (size_t i=0;i<type->field_count;i++) {
  const adamic_decode_field *f=&type->fields[i];
  const json_node *v=tuple?value->children[i]:field(value,f->name);
  if (v==NULL) {
   if (f->optional) { continue; }
   missing(error,path,f->name); adamic_release(o); return false;
  }
  adamic_string *where=path_field(path,f->name,tuple); adamic_value child={0};
  bool ok=validate(schema,f->node,v,where,&child,error); adamic_release(where);
  if (!ok) { adamic_release(o); return false; }
  ((const char **)o->shape->names)[written]=f->name;
  ((bool *)o->shape->references)[written]=reference(schema->nodes[f->node].representation);
  o->slots[written++]=child;
 }
 o->shape=canonical_shape(o->shape);
 out->reference=o; return true;
}
static adamic_string ok_kind=ADAMIC_STRING("Ok"), error_kind=ADAMIC_STRING("Error");
static const char *const ok_names[] = {"kind", "value"};
static const char *const error_names[] = {"kind", "message"};
static const bool references[]={true,true},number_references[]={true,false};
static const adamic_shape ok_shape = {2, ok_names, references, NULL};
static const adamic_shape number_shape = {2, ok_names, number_references, NULL};
static const adamic_shape error_shape = {2, error_names, references, NULL};
adamic_object *adamic_json_decode(const adamic_string *text,const adamic_decode_schema *schema) {
 json_reader reader={text,0,(size_t)adamic_string_length(text),NULL};
 json_node *tree=read_value(&reader,0);
 if (reader.error==NULL) { whitespace(&reader); }
 if (reader.error==NULL && reader.position!=reader.length) { failure(&reader,"expected end of input"); }
 adamic_string *error=NULL; adamic_value value={0};
 if (reader.error!=NULL) { error=syntax_message(&reader); }
 else {
  static adamic_string root=ADAMIC_STRING("$");
  (void)validate(schema,schema->root,tree,&root,&value,&error);
 }
 dispose(tree);
 if (error!=NULL) {
  adamic_object *result=adamic_object_new(&error_shape);
  result->slots[0].reference=&error_kind; result->slots[1].reference=error; return result;
 }
 int of=schema->nodes[schema->root].representation;
 adamic_object *result=adamic_object_new(reference(of)?&ok_shape:&number_shape);
 result->slots[0].reference=&ok_kind; result->slots[1]=value; return result;
}
