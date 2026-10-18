#include "regexp_compile_bytecode.h"
#ifdef ADAMIC_REGEXP_RUNTIME_COMPILER
#include <limits.h>
#include <stdlib.h>
#include <string.h>

typedef adamic_regex_parse_node regex_node;
typedef adamic_regex_parse_result regex_context;
typedef struct {
    regex_context *result;
    adamic_regex_program *program;
    adamic_regex_instruction *code;
    size_t count, capacity;
} regex_compiler;
static void regex_compile_refuse(regex_context *r, const char *reason) {
    if(r->status==0){r->status=4;r->reference_reason=reason;r->message=reason;r->message_length=strlen(reason);}
}
static size_t regex_compile_emit(regex_compiler *c, adamic_regex_instruction instruction) {
    if(c->result->status!=0)return 0;
    if(c->count >= INT_MAX){regex_compile_refuse(c->result,"native regexp instruction count exceeds int");return 0;}
    if(c->count==c->capacity) {
        size_t capacity=c->capacity==0 ? 16 : c->capacity*2;
        if(capacity<c->capacity || capacity>SIZE_MAX/sizeof(*c->code)){c->result->status=2;return 0;}
        adamic_regex_instruction *code=adamic_regex_parse_allocate(c->result,capacity*sizeof(*code));if(code==NULL)return 0;
        if(c->count!=0)memcpy(code,c->code,c->count*sizeof(*code));
        c->code=code;c->capacity=capacity;
    }
    c->code[c->count]=instruction;
    return c->count++;
}
static adamic_regex_instruction regex_compile_instruction(int op) {
    adamic_regex_instruction instruction={0};instruction.op=op;instruction.unbounded=true;return instruction;
}
static unsigned regex_compile_modified(unsigned flags, const regex_node *node) {
    unsigned mask=REGEX_FLAG_I|REGEX_FLAG_M|REGEX_FLAG_S;
    return (flags|node->enable)&~(node->disable&mask);
}
static int regex_compile_capture_order(const void *left, const void *right) {
    const regex_node *a=*(const regex_node *const *)left,*b=*(const regex_node *const *)right;
    int name=strcmp(a->name,b->name);if(name!=0)return name;
    return a->index==b->index ? 0 : a->index<b->index ? -1 : 1;
}
static void regex_compile_collect_names(const regex_node *node, const regex_node **names, size_t *count) {
    if(node==NULL)return;
    if(node->type==REGEX_PARSE_GROUP && node->kind==0 && node->name!=NULL)names[(*count)++]=node;
    regex_compile_collect_names(node->left,names,count);regex_compile_collect_names(node->right,names,count);
    for(const regex_node *child=node->children;child!=NULL;child=child->next)regex_compile_collect_names(child,names,count);
}
static bool regex_compile_names(regex_context *r, adamic_regex_program *p) {
    if(r->captures==0)return true;
    if(r->captures>SIZE_MAX/sizeof(regex_node *)){r->status=2;return false;}
    const regex_node **names=adamic_regex_parse_allocate(r,r->captures*sizeof(*names));if(names==NULL)return false;
    size_t count=0;regex_compile_collect_names(r->body,names,&count);if(count==0)return true;
    qsort(names,count,sizeof(*names),regex_compile_capture_order);
    adamic_regex_group *groups=adamic_regex_parse_allocate(r,count*sizeof(*groups));
    const char **shape_names=adamic_regex_parse_allocate(r,count*sizeof(*shape_names));
    bool *references=adamic_regex_parse_allocate(r,count*sizeof(*references));
    adamic_shape *shape=adamic_regex_parse_allocate(r,sizeof(*shape));
    if(groups==NULL||shape_names==NULL||references==NULL||shape==NULL)return false;
    size_t groups_count=0;
    for(size_t i=0;i<count;) {
        size_t end=i+1;while(end<count && strcmp(names[i]->name,names[end]->name)==0)end++;
        size_t *ids=adamic_regex_parse_allocate(r,(end-i)*sizeof(*ids));if(ids==NULL)return false;
        for(size_t j=i;j<end;j++)ids[j-i]=names[j]->index;
        groups[groups_count]=(adamic_regex_group){names[i]->name,ids,end-i};
        shape_names[groups_count]=names[i]->name;references[groups_count]=true;groups_count++;i=end;
    }
    *shape=(adamic_shape){groups_count,shape_names,references,NULL};p->groups=groups;p->group_count=groups_count;p->group_shape=shape;return true;
}
static bool regex_compile_bound(regex_context *r, const char *digits, uint64_t *bound) {
    *bound=0;if(digits==NULL)return true;
    for(const char *at=digits;*at!=0;at++) {
        unsigned digit=(unsigned)(*at-'0');
        if(*bound>(UINT64_MAX-digit)/10){regex_compile_refuse(r,"native regexp quantifier bounds above uint64 are not yet supported");return false;}
        *bound=*bound*10+digit;
    }
    return true;
}
static void regex_compile_clear(const regex_node *node, size_t *ids, size_t *count) {
    if(node==NULL)return;
    if(node->type==REGEX_PARSE_GROUP && node->kind==0)ids[(*count)++]=node->index;
    regex_compile_clear(node->left,ids,count);regex_compile_clear(node->right,ids,count);
    for(const regex_node *child=node->children;child!=NULL;child=child->next)regex_compile_clear(child,ids,count);
}
static bool regex_compile_node(regex_compiler *,const regex_node *,unsigned,int);
static bool regex_compile_disjunction(regex_compiler *c,const regex_node *node,unsigned flags,int direction) {
    size_t alternatives=0;for(const regex_node *a=node->children;a!=NULL;a=a->next)alternatives++;
    if(alternatives>SIZE_MAX/sizeof(size_t)){c->result->status=2;return false;}
    size_t *jumps=adamic_regex_parse_allocate(c->result,alternatives*sizeof(*jumps));if(jumps==NULL)return false;
    size_t jump_count=0;
    for(const regex_node *a=node->children;a!=NULL;a=a->next) {
        size_t split=SIZE_MAX;
        if(a->next!=NULL){adamic_regex_instruction i=regex_compile_instruction(2);i.x=(int)c->count+1;split=regex_compile_emit(c,i);}
        size_t count=0;for(const regex_node *term=a->children;term!=NULL;term=term->next)count++;
        if(count>SIZE_MAX/sizeof(regex_node *)){c->result->status=2;return false;}
        const regex_node **terms=adamic_regex_parse_allocate(c->result,count*sizeof(*terms));if(terms==NULL)return false;
        size_t k=0;for(const regex_node *term=a->children;term!=NULL;term=term->next)terms[k++]=term;
        for(size_t j=0;j<count;j++)if(!regex_compile_node(c,terms[direction<0 ? count-1-j : j],flags,direction))return false;
        if(split!=SIZE_MAX) {jumps[jump_count++]=regex_compile_emit(c,regex_compile_instruction(3));if(c->result->status!=0)return false;c->code[split].y=(int)c->count;}
    }
    for(size_t i=0;i<jump_count;i++)c->code[jumps[i]].x=(int)c->count;
    return c->result->status==0;
}
static bool regex_compile_node(regex_compiler *c,const regex_node *node,unsigned flags,int direction) {
    if(c->result->status!=0)return false;
    regex_context *r=c->result;
    adamic_regex_instruction base=regex_compile_instruction(0);base.direction=direction;base.flags=flags;
    switch(node->type) {
    case REGEX_PARSE_DISJUNCTION:return regex_compile_disjunction(c,node,flags,direction);
    case REGEX_PARSE_CHARACTER:case REGEX_PARSE_CLASS:case REGEX_PARSE_DOT: {
        adamic_regex_compile_set set={0};
        if(node->type==REGEX_PARSE_CHARACTER)set=adamic_regex_compile_character(r,node,flags);
        else if(node->type==REGEX_PARSE_CLASS){set=adamic_regex_compile_class(r,node->left,flags);if(node->negated)set=adamic_regex_compile_negate(r,set,flags);}
        else if(flags & REGEX_FLAG_S)adamic_regex_compile_add_range(r,&set,0,flags & REGEX_FLAG_U ? 0x10ffff : 0xffff);
        else {
            adamic_regex_compile_add_range(r,&set,0,9);adamic_regex_compile_add_range(r,&set,11,12);
            adamic_regex_compile_add_range(r,&set,14,0x2027);adamic_regex_compile_add_range(r,&set,0x202a,flags & REGEX_FLAG_U ? 0x10ffff : 0xffff);
        }
        if(r->status!=0)return false;
        base.op=1;base.ranges=set.ranges;base.range_count=set.range_count;base.strings=set.strings;base.string_count=set.string_count;regex_compile_emit(c,base);break;
    }
    case REGEX_PARSE_ASSERTION:base.op=5;base.assertion=node->kind;regex_compile_emit(c,base);break;
    case REGEX_PARSE_REFERENCE:
        base.op=7;
        if(node->name!=NULL){for(size_t i=0;i<c->program->group_count;i++)if(strcmp(c->program->groups[i].name,node->name)==0){base.ids=c->program->groups[i].captures;base.id_count=c->program->groups[i].count;break;}}
        else {size_t *id=adamic_regex_parse_allocate(r,sizeof(*id));if(id==NULL)return false;*id=node->index;base.ids=id;base.id_count=1;}
        regex_compile_emit(c,base);break;
    case REGEX_PARSE_GROUP: {
        unsigned modified=regex_compile_modified(flags,node);
        if(node->kind>=2) {
            int child_direction=node->kind==4||node->kind==5 ? -1 : 1;
            adamic_regex_program *sub=adamic_regex_parse_allocate(r,sizeof(*sub));if(sub==NULL)return false;
            sub->flags=modified;sub->captures=c->program->captures;sub->groups=c->program->groups;sub->group_count=c->program->group_count;sub->group_shape=c->program->group_shape;
            regex_compiler child={.result=r,.program=sub};if(!regex_compile_disjunction(&child,node->left,modified,child_direction))return false;
            regex_compile_emit(&child,regex_compile_instruction(0));sub->code=child.code;
            base.op=6;base.look=sub;base.negative=node->kind==3||node->kind==5;regex_compile_emit(c,base);
        } else {
            if(node->kind==0){adamic_regex_instruction save=regex_compile_instruction(4);save.x=(int)(2*node->index)+(direction<0);regex_compile_emit(c,save);}
            if(!regex_compile_disjunction(c,node->left,modified,direction))return false;
            if(node->kind==0){adamic_regex_instruction save=regex_compile_instruction(4);save.x=(int)(2*node->index)+(direction>0);regex_compile_emit(c,save);}
        }
        break;
    }
    case REGEX_PARSE_QUANTIFIER: {
        size_t id=c->program->repeats++;
        adamic_regex_instruction initialize=regex_compile_instruction(8);initialize.x=(int)id;regex_compile_emit(c,initialize);
        adamic_regex_instruction repeat=regex_compile_instruction(9);repeat.x=(int)id;repeat.greedy=node->greedy;repeat.unbounded=node->maximum==NULL;
        if(!regex_compile_bound(r,node->minimum,&repeat.minimum)||!regex_compile_bound(r,node->maximum,&repeat.maximum))return false;
        if(r->captures!=0){size_t *ids=adamic_regex_parse_allocate(r,r->captures*sizeof(*ids));if(ids==NULL)return false;size_t count=0;regex_compile_clear(node->left,ids,&count);repeat.ids=count==0 ? NULL : ids;repeat.id_count=count;}
        size_t choice=regex_compile_emit(c,repeat);
        if(!regex_compile_node(c,node->left,flags,direction))return false;
        adamic_regex_instruction end=regex_compile_instruction(10);end.x=(int)id;end.y=(int)choice;end.minimum=repeat.minimum;regex_compile_emit(c,end);
        if(r->status!=0)return false;c->code[choice].y=(int)c->count;break;
    }
    default:regex_compile_refuse(r,"unsupported regexp syntax");return false;
    }
    return r->status==0;
}
adamic_regex_program *adamic_regex_compile_bytecode(regex_context *result) {
    if(result->status!=0)return NULL;
    if(result->captures>INT_MAX/2){regex_compile_refuse(result,"native regexp capture count exceeds int");return NULL;}
    adamic_regex_program *program=adamic_regex_parse_allocate(result,sizeof(*program));if(program==NULL)return NULL;
    program->flags=result->flags;program->captures=result->captures;
    if(!regex_compile_names(result,program))return NULL;
    regex_compiler compiler={.result=result,.program=program};
    if(!regex_compile_disjunction(&compiler,result->body,result->flags,1))return NULL;
    regex_compile_emit(&compiler,regex_compile_instruction(0));if(result->status!=0)return NULL;
    program->code=compiler.code;return program;
}
#endif
