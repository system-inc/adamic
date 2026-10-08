// Linux/x86-64 scratch profiler. No production source or ownership change.
#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/syscall.h>
#include <time.h>
#include <ucontext.h>
#include <unistd.h>

static volatile sig_atomic_t scout_active;
static uintptr_t scout_pc[65536];
static uint64_t scout_hits[65536];
static uint64_t scout_samples, scout_overflow, scout_calls, scout_elapsed, scout_timed;
static uint64_t scout_entries[5];
static timer_t scout_timer;

static uint64_t scout_nanoseconds(void) {
 struct timespec t; if (clock_gettime(CLOCK_MONOTONIC, &t)) abort();
 return (uint64_t)t.tv_sec*1000000000+(uint64_t)t.tv_nsec;
}

static void scout_tick(int signal, siginfo_t *info, void *context) {
 (void)signal; (void)info;
 if (!scout_active) return;
 uintptr_t pc=(uintptr_t)((ucontext_t *)context)->uc_mcontext.gregs[REG_RIP];
 size_t i=(pc>>4)%65536;
 for (size_t n=0;n<65536;n++,i=(i+1)%65536) {
  if (!scout_pc[i] || scout_pc[i]==pc) { scout_pc[i]=pc; scout_hits[i]++; scout_samples++; return; }
 }
 scout_overflow++;
}

static void scout_finish(void) {
 if (getenv("SCOUT_PC")) timer_delete(scout_timer);
 fprintf(stderr,"SCOUT_WALL {\"calls\":%llu,\"sampled_ns\":%llu,\"timed_calls\":%llu,\"pc_samples\":%llu,\"overflow\":%llu,\"entries\":[%llu,%llu,%llu,%llu,%llu]}\n",
  (unsigned long long)scout_calls,(unsigned long long)scout_elapsed,(unsigned long long)scout_timed,
  (unsigned long long)scout_samples,(unsigned long long)scout_overflow,
  (unsigned long long)scout_entries[0],(unsigned long long)scout_entries[1],(unsigned long long)scout_entries[2],(unsigned long long)scout_entries[3],(unsigned long long)scout_entries[4]);
 const char *prefix=getenv("SCOUT_PC"); if (!prefix) return;
 char path[4096]; snprintf(path,sizeof path,"%s.pcs",prefix); FILE *out=fopen(path,"w"); if (!out) abort();
 for(size_t i=0;i<65536;i++) if(scout_hits[i]) fprintf(out,"%lx %llu\n",(unsigned long)scout_pc[i],(unsigned long long)scout_hits[i]);
 fclose(out);snprintf(path,sizeof path,"%s.maps",prefix);out=fopen(path,"w");FILE *in=fopen("/proc/self/maps","r");if(!out||!in)abort();
 char line[4096];while(fgets(line,sizeof line,in)) fputs(line,out);fclose(in);fclose(out);
}

static void scout_start(void) {
 if (atexit(scout_finish)) abort();
 if (!getenv("SCOUT_PC")) return;
 struct sigaction action={0};action.sa_sigaction=scout_tick;action.sa_flags=SA_SIGINFO|SA_RESTART|SA_ONSTACK;
 sigemptyset(&action.sa_mask);if(sigaction(SIGALRM,&action,NULL))abort();
 struct sigevent event={0};event.sigev_notify=SIGEV_THREAD_ID;event.sigev_signo=SIGALRM;event._sigev_un._tid=(pid_t)syscall(SYS_gettid);
 if(timer_create(CLOCK_MONOTONIC,&event,&scout_timer))abort();
 struct itimerspec spec={{0,1000000},{0,1000000}};if(timer_settime(scout_timer,0,&spec,NULL))abort();
}
