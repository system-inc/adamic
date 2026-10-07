#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

adamic_string *baseline_decode_utf8(const unsigned char *, size_t);
static unsigned read16(const unsigned char *bytes) { return bytes[0] | ((unsigned)bytes[1]<<8); }
int main(int argc, char **argv) {
 if(argc!=2) return 2;
 FILE *file=fopen(argv[1],"rb");if(file==NULL)return 2;
 unsigned char header[4],prefix[8192],oracle[24576];
 size_t records=0,cases=0;
 while(fread(header,1,4,file)==4) {
  size_t length=read16(header), expected=read16(header+2);
  if(length>sizeof prefix || expected>sizeof oracle || fread(prefix,1,length,file)!=length || fread(oracle,1,expected,file)!=expected)return 2;
  for(size_t run=0;run<=64;run++)for(size_t offset=0;offset<8;offset++)for(size_t tail=0;tail<2;tail++) {
   size_t suffix=tail?run:0, size=run+length+suffix;
   // The allocation ends exactly at input[length]. ASan sees any over-read, including word tails.
   unsigned char *allocation=malloc(offset+size+(size==0 && offset==0));if(allocation==NULL)abort();
   unsigned char *bytes=allocation+offset;
   memset(bytes,'A',run);memcpy(bytes+run,prefix,length);memset(bytes+run+length,'A',suffix);
   adamic_string *current=adamic_decode_utf8(bytes,size), *baseline=baseline_decode_utf8(bytes,size);
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
