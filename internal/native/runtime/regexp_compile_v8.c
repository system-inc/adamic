#include "regexp_compile_v8.h"
#ifdef ADAMIC_REGEXP_RUNTIME_COMPILER
#include <string.h>

typedef adamic_regex_parse_node regex_node;
typedef adamic_regex_parse_result regex_context;
typedef adamic_regex_compile_set regex_set;
typedef struct {regex_context *result;unsigned root_flags,parser_flags;bool later_alternative;} regex_v8;
static bool regex_v8_refuse(regex_context *r,const char *behavior,const char *section) {
    if(r->status!=0)return false;
    const char *prefix="RegExp refused: V8 ",*middle=" departs from ECMA-262 ",*suffix="; native and JavaScript must agree";
    size_t length=strlen(prefix)+strlen(behavior)+strlen(middle)+strlen(section)+strlen(suffix);
    char *message=adamic_regex_parse_allocate(r,length+1);if(message==NULL)return false;
    size_t position=0;const char *parts[]={prefix,behavior,middle,section,suffix};
    for(size_t i=0;i<sizeof(parts)/sizeof(parts[0]);i++){size_t count=strlen(parts[i]);memcpy(message+position,parts[i],count);position+=count;}
    message[position]=0;r->status=3;r->reference_reason=behavior;r->message=message;r->message_length=position;return false;
}
static bool regex_v8_modifiers(regex_context *r) {
    return regex_v8_refuse(r,"leaks or drops a scoped i modifier in a later Unicode class or word escape","22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers");
}
static unsigned regex_v8_modified(unsigned flags,const regex_node *node) {return (flags|node->enable)&~(node->disable&(REGEX_FLAG_I|REGEX_FLAG_M|REGEX_FLAG_S));}
static bool regex_v8_word(regex_context *r,const regex_node *node,bool uppercase_only) {
    if(node==NULL)return false;
    if(node->type==REGEX_PARSE_CLASS_CHARACTER) {
        const regex_node *character=node->left;
        if(character->kind!=2 || character->start+1>=r->length)return false;
        unsigned char byte=r->source[character->start+1];return byte=='W'||(!uppercase_only && byte=='w');
    }
    if(node->type==REGEX_PARSE_UNION)for(const regex_node *child=node->children;child!=NULL;child=child->next)if(regex_v8_word(r,child,uppercase_only))return true;
    if(node->type==REGEX_PARSE_NEGATION)return regex_v8_word(r,node->left,uppercase_only);
    return false;
}
static regex_set regex_v8_class(regex_context *r,const regex_node *node,unsigned flags) {
    regex_set out={0};if(r->status!=0)return out;
    switch(node->type) {
    case REGEX_PARSE_CLASS_STRING:return adamic_regex_compile_class(r,node,flags);
    case REGEX_PARSE_UNION:
        for(const regex_node *child=node->children;child!=NULL;child=child->next)out=adamic_regex_compile_union(r,out,regex_v8_class(r,child,flags));
        return out;
    case REGEX_PARSE_INTERSECTION:case REGEX_PARSE_SUBTRACTION: {
        regex_set a=regex_v8_class(r,node->left,flags),b=regex_v8_class(r,node->right,flags);
        return adamic_regex_compile_combine(r,a,b,node->type==REGEX_PARSE_INTERSECTION);
    }
    case REGEX_PARSE_NEGATION:return adamic_regex_compile_negate(r,regex_v8_class(r,node->left,flags),flags&~REGEX_FLAG_V);
    default:return adamic_regex_compile_matched(r,adamic_regex_compile_class(r,node,flags),flags);
    }
}
static regex_set regex_v8_unicode_class(regex_context *r,const regex_node *node,unsigned flags,unsigned parser_flags) {
    regex_set out={0};if(r->status!=0)return out;
    if(node->type==REGEX_PARSE_CLASS_CHARACTER) {
        unsigned operand_flags=flags;
        const regex_node *character=node->left;
        if(character->kind==2 && character->start+1<r->length && (r->source[character->start+1]=='w'||r->source[character->start+1]=='W'))operand_flags=parser_flags;
        return adamic_regex_compile_matched(r,adamic_regex_compile_character(r,character,operand_flags),operand_flags);
    }
    if(node->type==REGEX_PARSE_UNION) {
        for(const regex_node *child=node->children;child!=NULL;child=child->next)out=adamic_regex_compile_union(r,out,regex_v8_unicode_class(r,child,flags,parser_flags));
        return out;
    }
    return adamic_regex_compile_matched(r,adamic_regex_compile_class(r,node,flags),flags);
}
static bool regex_v8_visit(regex_v8 *v,const regex_node *node,unsigned flags) {
    regex_context *r=v->result;if(node==NULL || r->status!=0)return false;
    switch(node->type) {
    case REGEX_PARSE_DISJUNCTION: {
        bool outer=v->later_alternative;size_t index=0;
        for(const regex_node *a=node->children;a!=NULL;a=a->next,index++) {
            v->later_alternative=outer||index!=0;
            for(const regex_node *term=a->children;term!=NULL;term=term->next)if(!regex_v8_visit(v,term,flags)){v->later_alternative=outer;return false;}
        }
        v->later_alternative=outer;break;
    }
    case REGEX_PARSE_QUANTIFIER:return regex_v8_visit(v,node->left,flags);
    case REGEX_PARSE_GROUP:
        flags=regex_v8_modified(flags,node);v->parser_flags=flags;return regex_v8_visit(v,node->left,flags);
    case REGEX_PARSE_CLASS: {
        regex_set expected=adamic_regex_compile_class(r,node->left,flags);if(r->status!=0)return false;
        if(!(flags&REGEX_FLAG_U)&&(flags&REGEX_FLAG_I)&&!(v->root_flags&REGEX_FLAG_I)&&v->later_alternative&&node->negated) {
            regex_set actual=adamic_regex_compile_class(r,node->left,flags&~REGEX_FLAG_I);
            if(!adamic_regex_compile_same(adamic_regex_compile_matched(r,expected,flags),actual))return regex_v8_refuse(r,"drops scoped i on a later alternative's negated legacy class","22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers");
        }
        if(flags&REGEX_FLAG_V) {
            bool empty=false,long_string=false;
            for(size_t i=0;i<expected.string_count;i++){empty|=expected.strings[i].count==0;long_string|=expected.strings[i].count>1;}
            if(expected.range_count>0&&empty&&long_string)return regex_v8_refuse(r,"can repeat an earlier empty match and hang replacement for mixed-length \\q alternatives","22.2.2.7 CompileAtom and 22.2.6.11 RegExp.prototype [ Symbol.replace ] / AdvanceStringIndex");
            regex_set actual=regex_v8_class(r,node->left,v->parser_flags);
            regex_set strings={.strings=actual.strings,.string_count=actual.string_count,.string_capacity=actual.string_capacity};
            strings=adamic_regex_compile_fold(r,strings,flags);actual.strings=strings.strings;actual.string_count=strings.string_count;actual.string_capacity=strings.string_capacity;
            if(node->negated){expected=adamic_regex_compile_negate(r,expected,flags);actual=adamic_regex_compile_negate(r,actual,flags&~REGEX_FLAG_V);}
            if(!adamic_regex_compile_same(adamic_regex_compile_matched(r,expected,flags),actual)) {
                if((v->parser_flags&REGEX_FLAG_I)!=(flags&REGEX_FLAG_I))return regex_v8_modifiers(r);
                return regex_v8_refuse(r,"does not case-fold singleton \\q class strings under iv","22.2.2.9 CompileToCharSet and 22.2.2.10 CompileClassSetString");
            }
        } else if((flags&REGEX_FLAG_U)&&((v->parser_flags&REGEX_FLAG_I)!=(flags&REGEX_FLAG_I))&&regex_v8_word(r,node->left,!(v->parser_flags&REGEX_FLAG_I))) {
            regex_set actual=regex_v8_unicode_class(r,node->left,flags,v->parser_flags);actual=adamic_regex_compile_fold(r,actual,flags);
            if(node->negated){expected=adamic_regex_compile_negate(r,expected,flags);actual=adamic_regex_compile_negate(r,actual,flags);}
            if(!adamic_regex_compile_same(adamic_regex_compile_matched(r,expected,flags),adamic_regex_compile_matched(r,actual,flags)))return regex_v8_modifiers(r);
        }
        break;
    }
    case REGEX_PARSE_CHARACTER:
        if((flags&REGEX_FLAG_U)&&((v->parser_flags&REGEX_FLAG_I)!=(flags&REGEX_FLAG_I))) {
            if((flags&REGEX_FLAG_V)&&node->kind==3) {
                regex_set expected=adamic_regex_compile_character(r,node,flags),actual=adamic_regex_compile_character(r,node,v->parser_flags);
                if(!adamic_regex_compile_same(adamic_regex_compile_matched(r,expected,flags),adamic_regex_compile_matched(r,actual,v->parser_flags)))return regex_v8_modifiers(r);
            } else if(node->kind==2&&node->start+1<r->length) {
                unsigned char byte=r->source[node->start+1];if(byte=='W'||((v->parser_flags&REGEX_FLAG_I)&&(byte=='w'||byte=='W')))return regex_v8_modifiers(r);
            }
        }
        break;
    default:break;
    }
    return r->status==0;
}
bool adamic_regex_compile_v8_check(regex_context *result) {
    if(result->status!=0)return false;
    regex_v8 state={.result=result,.root_flags=result->flags,.parser_flags=result->flags};
    return regex_v8_visit(&state,result->body,result->flags);
}
adamic_regex_program *adamic_regex_compile_checked(regex_context *result) {
    adamic_regex_program *program=adamic_regex_compile_bytecode(result);
    if(program==NULL || !adamic_regex_compile_v8_check(result))return NULL;
    return program;
}
#endif
