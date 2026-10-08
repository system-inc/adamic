#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_string *baseline_decode_utf8(const unsigned char *, size_t);
static unsigned read16(const unsigned char *bytes) { return bytes[0] | ((unsigned)bytes[1]<<8); }
// Exercise the public indexing path before metadata assertions: a false ASCII flag
// returns UTF-8 bytes here instead of JavaScript UTF-16 code units.
static int indexing_probe(void) {
 const unsigned char input[] = {0xc3,0xa9,0xf0,0x9f,0x98,0x80};
 adamic_string *string=adamic_decode_utf8(input,sizeof input);
 const unsigned expected[] = {0xe9,0xd83d,0xde00};
 for(size_t index=0;index<3;index++) {
  double actual=adamic_string_char_code_at(string,(double)index);
  if(actual!=expected[index]) {
   fprintf(stderr,"decode indexing mismatch index=%zu actual=%.0f expected=%u\n",index,actual,expected[index]);
   adamic_release(string);return 1;
  }
 }
 adamic_release(string);return 0;
}
int main(int argc, char **argv) {
 if(argc!=2) return 2;
 if(indexing_probe()!=0)return 1;
 FILE *file=fopen(argv[1],"rb");if(file==NULL)return 2;
 unsigned char header[4],prefix[8192],oracle[24576];
 size_t records=0,cases=0;
 while(fread(header,1,4,file)==4) {
  size_t length=read16(header), expected=read16(header+2);
  if(length>sizeof prefix || expected>sizeof oracle || fread(prefix,1,length,file)!=length || fread(oracle,1,expected,file)!=expected)return 2;
  // Node's decoded UTF-8 oracle supplies the expected UTF-16 count independently.
  size_t oracle_units=0;for(size_t k=0;k<expected;k++)if(oracle[k]<0x80 || oracle[k]>=0xc0)oracle_units+=oracle[k]>=0xf0?2:1;
  bool ascii=true;for(size_t k=0;k<length;k++)if(prefix[k]>=0x80)ascii=false;
  for(size_t run=0;run<=64;run++)for(size_t offset=0;offset<8;offset++)for(size_t tail=0;tail<2;tail++) {
   size_t suffix=tail?run:0, size=run+length+suffix;
   // The allocation ends exactly at input[length]. ASan sees any over-read, including word tails.
   unsigned char *allocation=malloc(offset+size+(size==0 && offset==0));if(allocation==NULL)abort();
   unsigned char *bytes=allocation+offset;
   memset(bytes,'A',run);memcpy(bytes+run,prefix,length);memset(bytes+run+length,'A',suffix);
   adamic_string *current=adamic_decode_utf8(bytes,size), *baseline=baseline_decode_utf8(bytes,size);
   bool flag=current->units==current->length+1;
   bool cache=flag==ascii && current->units==run+oracle_units+suffix+1;
   if(!cache){fprintf(stderr,"decode cache mismatch record=%zu run=%zu offset=%zu ascii=%d units=%zu bytes=%zu\n",records,run,offset,ascii,current->units,current->length);adamic_release(current);adamic_release(baseline);free(allocation);fclose(file);return 1;}
   bool equal=current->length==run+expected+suffix && baseline->length==current->length;
   if(equal) equal=memcmp(current->bytes,baseline->bytes,current->length)==0 && memcmp(current->bytes+run,oracle,expected)==0;
   for(size_t k=0;equal && k<run;k++)equal=current->bytes[k]=='A';
   for(size_t k=0;equal && k<suffix;k++)equal=current->bytes[run+expected+k]=='A';
   if(!equal){fprintf(stderr,"decode mismatch record=%zu run=%zu offset=%zu tail=%zu lead=%02x\n",records,run,offset,tail,length?prefix[0]:0);adamic_release(current);adamic_release(baseline);free(allocation);fclose(file);return 1;}
   adamic_release(current);adamic_release(baseline);free(allocation);cases++;
  }
  records++;
 }
 if(ferror(file))return 2;
 fclose(file);printf("prefixes=%zu cases=%zu Node and baseline identical\n",records,cases);return 0;
}
