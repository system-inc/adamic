#define _GNU_SOURCE
#include <dlfcn.h>
#include <stdarg.h>
#include <sys/ioctl.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
int ioctl(int fd, unsigned long request, ...) {
 va_list args; va_start(args, request); void *arg=va_arg(args,void *); va_end(args);
 static int (*original)(int,unsigned long,...);
 if (!original) {void *symbol=dlsym(RTLD_NEXT,"ioctl"); memcpy(&original,&symbol,sizeof original);}
 if(request==TIOCGWINSZ && getenv("COLUMNS_IOCTL_LOG")) {
  int log=open(getenv("COLUMNS_IOCTL_LOG"),O_WRONLY|O_CREAT|O_APPEND,0600);
  if(log>=0) { (void)write(log,"w",1);close(log); }
 }
 if(request==TIOCGWINSZ && getenv("COLUMNS_FAIL_UNTIL") && access(getenv("COLUMNS_FAIL_UNTIL"),F_OK)!=0) {errno=EIO;return -1;}
 return original(fd,request,arg);
}
