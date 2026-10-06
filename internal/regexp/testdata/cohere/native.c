/* Observation-only harness. Node's expected values are never compiled into C. */
static void json_name(const char *name) {
 putchar('"');
 for(const unsigned char *c=(const unsigned char *)name;*c;c++) {
  if(*c=='"'||*c=='\\') {putchar('\\');putchar(*c);}
  else if(*c<32) printf("\\u%04x",*c);
  else putchar(*c);
 }
 putchar('"');
}
static void print_value(adamic_string *value) {
 if(value==NULL) {fputs("null",stdout);return;}
 putchar('[');size_t length=(size_t)adamic_string_length(value);
 for(size_t k=0;k<length;k++) {if(k) putchar(',');printf("%.0f",adamic_string_char_code_at(value,(double)k));}
 putchar(']');
}
static void print_pair(adamic_object *pair) {
 if(pair==NULL) fputs("null",stdout);
 else printf("[%.0f,%.0f]",pair->slots[0].number,pair->slots[1].number);
}
static void print_groups(adamic_object *groups,bool indices) {
 if(groups==NULL) {fputs("null",stdout);return;}
 putchar('{');
 for(size_t k=0;k<groups->shape->count;k++) {
  if(k) putchar(',');json_name(groups->shape->names[k]);putchar(':');
  if(indices) print_pair(groups->slots[k].reference);else print_value(groups->slots[k].reference);
 }
 putchar('}');
}
int main(int argc,char **argv) {
 adamic_start(argc,argv);
 bool bounded=argc<2 || strcmp(argv[1],"bounded")==0;
 adamic_regex_set_step_limit(bounded?10000000:0);
 adamic_regex_set_regular_mode(argc>1 && strcmp(argv[1],"vm")==0?0:1);
 adamic_object *regex=NULL;size_t series=SIZE_MAX;
 for(size_t index=0;index<sizeof probes/sizeof probes[0];index++) {
  const probe *p=&probes[index];
  if(series!=p->series) {
   adamic_release(regex);regex=adamic_regex_new(p->program,&adamic_string_empty,&adamic_string_empty);
   series=p->series;regex->slots[1].number=(double)p->last;
  }
  double *codes=malloc((p->length+1)*sizeof *codes);if(codes==NULL) abort();
  for(size_t k=0;k<p->length;k++) codes[k]=input_units[p->offset+k];
  adamic_string *input=adamic_string_from_char_codes(p->length,codes);free(codes);
  adamic_array *match=adamic_regex_exec(regex,input);
  printf("{\"case\":%zu,\"result\":{\"index\":",p->index);
  if(match==NULL) fputs("null,\"captures\":null,\"values\":null,\"groups\":null,\"groupValues\":null",stdout);
  else {
   printf("%.0f,\"captures\":[",match->properties->slots[0].number);
   adamic_array *indices=match->properties->slots[3].reference;
   for(size_t k=0;k<indices->length;k++) {if(k) putchar(',');print_pair(indices->elements[k].reference);}
   fputs("],\"values\":[",stdout);
   for(size_t k=0;k<match->length;k++) {if(k) putchar(',');print_value(match->elements[k].reference);}
   fputs("],\"groups\":",stdout);print_groups(indices->properties->slots[2].reference,true);
   fputs(",\"groupValues\":",stdout);print_groups(match->properties->slots[2].reference,false);
  }
  printf(",\"lastIndex\":%.0f}}\n",regex->slots[1].number);
  if(match!=NULL && (p->program->flags&(8|16))) {
   adamic_array *indices=match->properties->slots[3].reference;
   adamic_object *whole=indices->elements[0].reference;
   if(whole->slots[0].number==whole->slots[1].number) {
    size_t at=(size_t)regex->slots[1].number;
    size_t next=at+1;
    if((p->program->flags&4) && at+1<p->length && input_units[p->offset+at]>=0xd800 && input_units[p->offset+at]<=0xdbff && input_units[p->offset+at+1]>=0xdc00 && input_units[p->offset+at+1]<=0xdfff) next++;
    regex->slots[1].number=(double)next;
   }
  }
  adamic_release(match);adamic_release(input);
 }
 adamic_release(regex);return 0;
}
