#include "regexp_compile_sets.h"
#ifdef ADAMIC_REGEXP_RUNTIME_COMPILER
#include "regexp_fold.h"
#include <stdlib.h>
#include <string.h>

typedef adamic_regex_compile_set regex_set;
typedef adamic_regex_parse_result regex_context;
typedef adamic_regex_parse_node regex_node;
static uint32_t regex_compile_max(unsigned flags) { return flags & REGEX_FLAG_U ? 0x10ffff : 0xffff; }
static void *regex_compile_grow(regex_context *r, const void *old, size_t count, size_t *capacity, size_t width) {
    size_t next = *capacity == 0 ? 8 : *capacity;
    if (next > SIZE_MAX / 2 || next * 2 > SIZE_MAX / width) {r->status = 2; return NULL;}
    next *= 2;
    void *memory = adamic_regex_parse_allocate(r, next * width);
    if (memory == NULL) return NULL;
    if (count != 0) memcpy(memory, old, count * width);
    *capacity = next;
    return memory;
}
void adamic_regex_compile_add_range(regex_context *r, regex_set *set, uint32_t first, uint32_t last) {
    if (first > last || r->status != 0) return;
    if (set->range_count == set->range_capacity) {
        adamic_regex_range *ranges = regex_compile_grow(r, set->ranges, set->range_count, &set->range_capacity, sizeof(*ranges));
        if (ranges == NULL) return;
        set->ranges = ranges;
    }
    set->ranges[set->range_count++] = (adamic_regex_range){first,last};
}
static void regex_compile_add_string(regex_context *r, regex_set *set, const uint32_t *points, size_t count) {
    if (r->status != 0) return;
    if (set->string_count == set->string_capacity) {
        adamic_regex_text *strings = regex_compile_grow(r, set->strings, set->string_count, &set->string_capacity, sizeof(*strings));
        if (strings == NULL) return;
        set->strings = strings;
    }
    set->strings[set->string_count++] = (adamic_regex_text){points,count};
}
static int regex_compile_range_order(const void *left, const void *right) {
    const adamic_regex_range *a = left, *b = right;
    if (a->first != b->first) return a->first < b->first ? -1 : 1;
    return a->last == b->last ? 0 : a->last < b->last ? -1 : 1;
}
static regex_set regex_compile_normalize(regex_set set) {
    if (set.range_count == 0) return set;
    qsort(set.ranges, set.range_count, sizeof(*set.ranges), regex_compile_range_order);
    size_t count = 0;
    for (size_t i = 0; i < set.range_count; i++) {
        adamic_regex_range range = set.ranges[i];
        if (count != 0 && range.first <= set.ranges[count-1].last + 1) {
            if (range.last > set.ranges[count-1].last) set.ranges[count-1].last = range.last;
        } else set.ranges[count++] = range;
    }
    set.range_count = count;
    return set;
}
static bool regex_compile_contains(regex_set set, uint32_t point) {
    size_t low = 0, high = set.range_count;
    while (low < high) {
        size_t middle = low + (high-low)/2;
        if (set.ranges[middle].last < point) low = middle+1; else high = middle;
    }
    return low < set.range_count && set.ranges[low].first <= point;
}
static regex_set regex_compile_subtract(regex_context *r, regex_set a, regex_set b) {
    regex_set out = {0}; size_t j = 0;
    for (size_t i = 0; i < a.range_count; i++) {
        uint32_t first = a.ranges[i].first, last = a.ranges[i].last;
        while (j < b.range_count && b.ranges[j].last < first) j++;
        for (size_t k = j; k < b.range_count && b.ranges[k].first <= last; k++) {
            if (b.ranges[k].first > first) adamic_regex_compile_add_range(r,&out,first,b.ranges[k].first-1);
            if (b.ranges[k].last >= first) first = b.ranges[k].last+1;
            if (first > last) break;
        }
        if (first <= last) adamic_regex_compile_add_range(r,&out,first,last);
    }
    return out;
}
static regex_set regex_compile_intersect(regex_context *r, regex_set a, regex_set b) {
    regex_set out = {0}; size_t i = 0, j = 0;
    while (i < a.range_count && j < b.range_count) {
        uint32_t first = a.ranges[i].first > b.ranges[j].first ? a.ranges[i].first : b.ranges[j].first;
        uint32_t last = a.ranges[i].last < b.ranges[j].last ? a.ranges[i].last : b.ranges[j].last;
        if (first <= last) adamic_regex_compile_add_range(r,&out,first,last);
        if (a.ranges[i].last < b.ranges[j].last) i++; else j++;
    }
    return out;
}
static uint32_t regex_compile_canonical(uint32_t point, unsigned flags) {
    if (!(flags & REGEX_FLAG_I)) return point;
    const uint32_t (*table)[2] = flags & REGEX_FLAG_U ? regex_unicode_fold : regex_legacy_fold;
    size_t low = 0, high = flags & REGEX_FLAG_U ? sizeof(regex_unicode_fold)/sizeof(regex_unicode_fold[0]) : sizeof(regex_legacy_fold)/sizeof(regex_legacy_fold[0]);
    size_t count = high;
    while (low < high) {size_t middle = low+(high-low)/2; if(table[middle][0]<point)low=middle+1;else high=middle;}
    return low < count && table[low][0] == point ? table[low][1] : point;
}
adamic_regex_compile_set adamic_regex_compile_fold(regex_context *r, regex_set set, unsigned flags) {
    regex_set strings = {0};
    for (size_t i = 0; i < set.string_count; i++) {
        adamic_regex_text text = set.strings[i];
        if (text.count == 1) adamic_regex_compile_add_range(r,&set,text.points[0],text.points[0]);
        else regex_compile_add_string(r,&strings,text.points,text.count);
    }
    set.strings = strings.strings; set.string_count = strings.string_count; set.string_capacity = strings.string_capacity;
    set = regex_compile_normalize(set);
    if (!(flags & REGEX_FLAG_I) || r->status != 0) return set;
    const uint32_t (*table)[2] = flags & REGEX_FLAG_U ? regex_unicode_fold : regex_legacy_fold;
    size_t count = flags & REGEX_FLAG_U ? sizeof(regex_unicode_fold)/sizeof(regex_unicode_fold[0]) : sizeof(regex_legacy_fold)/sizeof(regex_legacy_fold[0]);
    regex_set sources = {0}, targets = {0};
    for (size_t i = 0; i < count; i++) {
        if (regex_compile_contains(set,table[i][0])) {
            adamic_regex_compile_add_range(r,&sources,table[i][0],table[i][0]);
            adamic_regex_compile_add_range(r,&targets,table[i][1],table[i][1]);
        }
    }
    regex_set changed = regex_compile_subtract(r,set,sources);
    for (size_t i = 0; i < targets.range_count; i++) adamic_regex_compile_add_range(r,&changed,targets.ranges[i].first,targets.ranges[i].last);
    changed = regex_compile_normalize(changed);
    for (size_t i = 0; i < set.string_count; i++) {
        adamic_regex_text text = set.strings[i];
        if (text.count > SIZE_MAX/sizeof(uint32_t)) {r->status = 2; return changed;}
        uint32_t *points = adamic_regex_parse_allocate(r,text.count*sizeof(*points));
        if (points == NULL) return changed;
        for (size_t j = 0; j < text.count; j++) points[j] = regex_compile_canonical(text.points[j],flags);
        regex_compile_add_string(r,&changed,points,text.count);
    }
    return changed;
}
adamic_regex_compile_set adamic_regex_compile_negate(regex_context *r, regex_set set, unsigned flags) {
    regex_set universe = {0}; adamic_regex_compile_add_range(r,&universe,0,regex_compile_max(flags));
    if (flags & REGEX_FLAG_V) universe = adamic_regex_compile_fold(r,universe,flags);
    return regex_compile_subtract(r,universe,set);
}
adamic_regex_compile_set adamic_regex_compile_character(regex_context *r, const regex_node *node, unsigned flags) {
    regex_set set = {0};
    unsigned char escape = node->start + 1 < r->length ? r->source[node->start+1] : 0;
    if (node->kind == 3) {
        const adamic_regex_compile_property *property = adamic_regex_compile_lookup_property((const unsigned char *)node->name,strlen(node->name),(flags & REGEX_FLAG_V)!=0);
        if (property == NULL) {r->status=4; r->reference_reason="unknown Unicode property"; return set;}
        for (size_t i=0;i<property->range_count;i++) adamic_regex_compile_add_range(r,&set,property->ranges[i].first,property->ranges[i].last);
        for (size_t i=0;i<property->string_count;i++) regex_compile_add_string(r,&set,property->strings[i].points,property->strings[i].count);
        set=regex_compile_normalize(set);
        if (escape=='P') {
            if(flags & REGEX_FLAG_V) return adamic_regex_compile_negate(r,adamic_regex_compile_fold(r,set,flags),flags);
            set=adamic_regex_compile_negate(r,set,flags);
        }
    } else if (node->kind == 2) {
        unsigned char lower=escape>='A'&&escape<='Z' ? escape+32 : escape;
        if(lower=='d')adamic_regex_compile_add_range(r,&set,'0','9');
        else if(lower=='w') {
            adamic_regex_compile_add_range(r,&set,'0','9'); adamic_regex_compile_add_range(r,&set,'A','Z');
            adamic_regex_compile_add_range(r,&set,'_','_'); adamic_regex_compile_add_range(r,&set,'a','z');
        } else if(lower=='s') {
            const adamic_regex_range spaces[]={{9,13},{32,32},{0xa0,0xa0},{0x1680,0x1680},{0x2000,0x200a},{0x2028,0x2029},{0x202f,0x202f},{0x205f,0x205f},{0x3000,0x3000},{0xfeff,0xfeff}};
            for(size_t i=0;i<sizeof(spaces)/sizeof(spaces[0]);i++)adamic_regex_compile_add_range(r,&set,spaces[i].first,spaces[i].last);
        }
        set=adamic_regex_compile_fold(r,set,flags);
        if(escape>='A'&&escape<='Z') {
            regex_set universe={0};adamic_regex_compile_add_range(r,&universe,0,regex_compile_max(flags));
            set=regex_compile_subtract(r,adamic_regex_compile_fold(r,universe,flags),set);
        }
        return set;
    } else {
        uint32_t value=node->value;
        if(value>0xffff && !(flags & REGEX_FLAG_U)) {
            uint32_t high=0xd800+(value-0x10000)/0x400, low=0xdc00+(value-0x10000)%0x400;
            adamic_regex_compile_add_range(r,&set,high,high);adamic_regex_compile_add_range(r,&set,low,low);
        } else adamic_regex_compile_add_range(r,&set,value,value);
    }
    return adamic_regex_compile_fold(r,set,flags);
}
static regex_set regex_compile_combine(regex_context *r, regex_set a, regex_set b, bool intersection) {
    regex_set out=intersection ? regex_compile_intersect(r,a,b) : regex_compile_subtract(r,a,b);
    for(size_t i=0;i<a.string_count;i++) {
        bool contains=false;
        for(size_t j=0;j<b.string_count;j++) if(a.strings[i].count==b.strings[j].count && (a.strings[i].count==0 || memcmp(a.strings[i].points,b.strings[j].points,a.strings[i].count*sizeof(uint32_t))==0)) {contains=true;break;}
        if(contains==intersection)regex_compile_add_string(r,&out,a.strings[i].points,a.strings[i].count);
    }
    return out;
}
adamic_regex_compile_set adamic_regex_compile_class(regex_context *r, const regex_node *node, unsigned flags) {
    regex_set out={0};if(node==NULL || r->status!=0)return out;
    switch(node->type) {
    case REGEX_PARSE_CLASS_CHARACTER:return adamic_regex_compile_character(r,node->left,flags);
    case REGEX_PARSE_RANGE:
        adamic_regex_compile_add_range(r,&out,node->left->value,node->right->value);return adamic_regex_compile_fold(r,out,flags);
    case REGEX_PARSE_NEGATION:return adamic_regex_compile_negate(r,adamic_regex_compile_class(r,node->left,flags),flags);
    case REGEX_PARSE_CLASS_STRING:
        for(const regex_node *a=node->children;a!=NULL;a=a->next) {
            size_t count=0;for(const regex_node *c=a->children;c!=NULL;c=c->next)count++;
            if(count>SIZE_MAX/sizeof(uint32_t)){r->status=2;return out;}
            uint32_t *points=adamic_regex_parse_allocate(r,count*sizeof(*points));if(points==NULL)return out;
            size_t i=0;for(const regex_node *c=a->children;c!=NULL;c=c->next)points[i++]=c->value;
            if(count==1)adamic_regex_compile_add_range(r,&out,points[0],points[0]);else regex_compile_add_string(r,&out,points,count);
        }
        return adamic_regex_compile_fold(r,regex_compile_normalize(out),flags);
    case REGEX_PARSE_UNION:
        for(const regex_node *child=node->children;child!=NULL;child=child->next) {
            regex_set set=adamic_regex_compile_class(r,child,flags);
            for(size_t i=0;i<set.range_count;i++)adamic_regex_compile_add_range(r,&out,set.ranges[i].first,set.ranges[i].last);
            for(size_t i=0;i<set.string_count;i++)regex_compile_add_string(r,&out,set.strings[i].points,set.strings[i].count);
        }
        return regex_compile_normalize(out);
    case REGEX_PARSE_INTERSECTION:case REGEX_PARSE_SUBTRACTION: {
        regex_set a=adamic_regex_compile_class(r,node->left,flags),b=adamic_regex_compile_class(r,node->right,flags);
        return regex_compile_combine(r,a,b,node->type==REGEX_PARSE_INTERSECTION);
    }
    default:r->status=4;r->reference_reason="unsupported class syntax";return out;
    }
}
adamic_regex_compile_set adamic_regex_compile_union(regex_context *r, regex_set a, regex_set b) {
    regex_set out={0};
    for(size_t i=0;i<a.range_count;i++)adamic_regex_compile_add_range(r,&out,a.ranges[i].first,a.ranges[i].last);
    for(size_t i=0;i<b.range_count;i++)adamic_regex_compile_add_range(r,&out,b.ranges[i].first,b.ranges[i].last);
    for(size_t i=0;i<a.string_count;i++)regex_compile_add_string(r,&out,a.strings[i].points,a.strings[i].count);
    for(size_t i=0;i<b.string_count;i++)regex_compile_add_string(r,&out,b.strings[i].points,b.strings[i].count);
    return regex_compile_normalize(out);
}
adamic_regex_compile_set adamic_regex_compile_combine(regex_context *r, regex_set a, regex_set b, bool intersection) {
    return regex_compile_combine(r,a,b,intersection);
}
adamic_regex_compile_set adamic_regex_compile_matched(regex_context *r, regex_set set, unsigned flags) {
    regex_set out=adamic_regex_compile_union(r,set,(regex_set){0});
    if(!(flags & REGEX_FLAG_I))return out;
    const uint32_t (*table)[2]=flags & REGEX_FLAG_U ? regex_unicode_fold : regex_legacy_fold;
    size_t count=flags & REGEX_FLAG_U ? sizeof(regex_unicode_fold)/sizeof(regex_unicode_fold[0]) : sizeof(regex_legacy_fold)/sizeof(regex_legacy_fold[0]);
    for(size_t i=0;i<count;i++)if(regex_compile_contains(set,table[i][1]))adamic_regex_compile_add_range(r,&out,table[i][0],table[i][0]);
    return regex_compile_normalize(out);
}
bool adamic_regex_compile_same(regex_set a, regex_set b) {
    if(a.range_count!=b.range_count || a.string_count!=b.string_count)return false;
    for(size_t i=0;i<a.range_count;i++)if(a.ranges[i].first!=b.ranges[i].first || a.ranges[i].last!=b.ranges[i].last)return false;
    for(size_t i=0;i<a.string_count;i++) {
        bool found=false;
        for(size_t j=0;j<b.string_count;j++)if(a.strings[i].count==b.strings[j].count && (a.strings[i].count==0 || memcmp(a.strings[i].points,b.strings[j].points,a.strings[i].count*sizeof(uint32_t))==0)){found=true;break;}
        if(!found)return false;
    }
    return true;
}
#endif
