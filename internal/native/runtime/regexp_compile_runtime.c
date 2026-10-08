#include "adamic.h"
#include "regexp_compile_v8.h"
#ifdef ADAMIC_REGEXP_RUNTIME_COMPILER
#include <string.h>
/* Each arena block is owned by a zero-length raw array, whose normal destructor
 * frees its buffer. A reference array owns all blocks; no cache or GC is used. */
static void regex_runtime_take(void *owner,void *block) {
    adamic_array *raw=adamic_array_new(0,false);raw->elements=block;
    adamic_array_push(owner,(adamic_value){.reference=raw});
}
static adamic_string *regex_runtime_source(adamic_string *pattern,unsigned flags) {
    if(pattern->length==0){adamic_string *empty=adamic_string_allocate(4);memcpy((char *)empty->bytes,"(?:)",4);return empty;}
    adamic_string *source=adamic_string_allocate(pattern->length*6);
    char *out=(char *)source->bytes;size_t count=0,depth=0;bool escaped=false;
    for(size_t i=0;i<pattern->length;i++){
        unsigned char byte=(unsigned char)pattern->bytes[i];const char *escape=NULL;
        if(byte=='\n')escape="n";else if(byte=='\r')escape="r";
        else if(byte==0xe2&&i+2<pattern->length&&(unsigned char)pattern->bytes[i+1]==0x80){if((unsigned char)pattern->bytes[i+2]==0xa8)escape="u2028";else if((unsigned char)pattern->bytes[i+2]==0xa9)escape="u2029";if(escape!=NULL)i+=2;}
        if(escape!=NULL){if(!escaped)out[count++]='\\';size_t n=strlen(escape);memcpy(out+count,escape,n);count+=n;escaped=false;continue;}
        if(!escaped){if(byte=='['&&((flags&REGEX_FLAG_V)||depth==0))depth++;if(byte==']'&&depth>0)depth--;}
        if(byte=='/'&&!escaped&&depth==0)out[count++]='\\';out[count++]=(char)byte;escaped=byte=='\\'?!escaped:false;
    }
    source->length=count;return source;
}
adamic_object *adamic_regex_compile_new(adamic_string *pattern,adamic_string *flags) {
    if(pattern==NULL)pattern=&adamic_string_empty;if(flags==NULL)flags=&adamic_string_empty;
    adamic_regex_parse_result result;adamic_regex_parse_native((const unsigned char *)pattern->bytes,pattern->length,(const unsigned char *)flags->bytes,flags->length,&result);
    adamic_regex_program *program=result.status==0?adamic_regex_compile_checked(&result):NULL;
    if(program==NULL){
        const char *message=result.message;if(message==NULL)message=result.reference_reason;if(message==NULL)message="RegExp runtime compiler: out of memory";
        size_t length=result.message!=NULL?result.message_length:strlen(message);
        adamic_string *text=adamic_string_allocate(length);memcpy((char *)text->bytes,message,length);
        adamic_object *error=adamic_error_new(text);adamic_release(text);
        if(result.status==1||result.status==3){adamic_string *name=adamic_string_allocate(11);memcpy((char *)name->bytes,"SyntaxError",11);adamic_release(error->slots[0].reference);error->slots[0].reference=name;}
        adamic_regex_parse_free(&result);adamic_thrown=error;return NULL;
    }
    if(program->group_count!=0){
        size_t n=program->group_count;adamic_shape *shape=adamic_regex_parse_allocate(&result,sizeof(*shape));
        const char **names=adamic_regex_parse_allocate(&result,(n+1)*sizeof(*names));bool *references=adamic_regex_parse_allocate(&result,(n+1)*sizeof(*references));
        if(shape==NULL||names==NULL||references==NULL)adamic_panic("out of memory",13);
        for(size_t i=0;i<n;i++){names[i]=program->group_shape->names[i];references[i]=true;}names[n]="#compiler";references[n]=true;
        *shape=(adamic_shape){n+1,names,references,NULL};program->group_shape=shape;
    }
    adamic_string *source=regex_runtime_source(pattern,program->flags);
    char ordered[8];size_t count=0;const char *letters="dgimsuvy";const unsigned bits[]={32,8,1,2,128,4,64,16};
    for(size_t i=0;i<8;i++)if((program->flags&bits[i])&&(letters[i]!='u'||!(program->flags&REGEX_FLAG_V)))ordered[count++]=letters[i];
    adamic_string *flag_string=adamic_string_allocate(count);memcpy((char *)flag_string->bytes,ordered,count);
    adamic_array *storage=adamic_array_new(0,true);adamic_regex_parse_take_memory(&result,regex_runtime_take,storage);
    adamic_object *regex=adamic_regex_new_owned(program,source,flag_string,storage);
    adamic_release(storage);adamic_release(source);adamic_release(flag_string);return regex;
}
#endif
