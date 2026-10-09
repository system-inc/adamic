#define _GNU_SOURCE
#include <sys/wait.h>
#include <sys/resource.h>
#include <time.h>
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
int main(int argc,char **argv){
 if(argc<3)return 125;
 struct timespec a,b; clock_gettime(CLOCK_MONOTONIC,&a);
 pid_t p=fork(); if(p<0)return 125;
 if(p==0){execvp(argv[2],argv+2);perror("execvp");_exit(127);}
 int status; struct rusage usage;
 while(wait4(p,&status,0,&usage)<0){if(errno!=EINTR)return 125;}
 clock_gettime(CLOCK_MONOTONIC,&b);
 FILE *out=fopen(argv[1],"w");if(!out)return 125;
 fprintf(out,"max_rss_kib=%ld\nwall_seconds=%.6f\nexit_code=%d\nsignal=%d\ncommand=",usage.ru_maxrss,(b.tv_sec-a.tv_sec)+(b.tv_nsec-a.tv_nsec)/1e9,WIFEXITED(status)?WEXITSTATUS(status):-1,WIFSIGNALED(status)?WTERMSIG(status):0);
 for(int i=2;i<argc;i++)fprintf(out,"%s%s",i==2?"":" ",argv[i]);
 fputc('\n',out);fclose(out);
 return WIFEXITED(status)?WEXITSTATUS(status):128+WTERMSIG(status);
}
