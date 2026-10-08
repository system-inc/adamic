#define _POSIX_C_SOURCE 200809L
#include "host_runtime.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>
static void log_bool(bool value) { puts(value ? "true" : "false"); }
static bool equal(const adamic_string *value, const char *text) {
 return value && value->length == strlen(text) && memcmp(value->bytes, text, value->length) == 0;
}
static void print_text(adamic_string *value) {
 if (!value) abort();
 printf("%.*s\n", (int)value->length, value->bytes);
}
static void forbidden_cleanup(void) { puts("finally"); }
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 const char *mode = argv[1];
 if (strcmp(mode,"output") == 0) {
  static adamic_string first = ADAMIC_STRING("Version 6.0.3\n");
  static adamic_string text = ADAMIC_STRING("h\xc3\xa9llo \xf0\x9f\x8c\x8d\0\xed\xa0\x80\n");
  static adamic_string error = ADAMIC_STRING("diagnostic\n");
  bool (*write)(enum adamic_stream, const adamic_string *) = adamic_write_raw;
  if (!adamic_write_raw(adamic_stdout,&first) || !write(adamic_stdout,&text) || !write(adamic_stderr,&error)) return 3;
 } else if (strcmp(mode,"exit") == 0) {
  adamic_host_set_exit_code(true,7);
  static adamic_string a = ADAMIC_STRING("retained "), b = ADAMIC_STRING("at exit");
  adamic_string *held = adamic_string_concat(2,(adamic_string *const[]){&a,&b});
  adamic_retain(held); // Explicit termination must neither free nor release this value.
  atexit(forbidden_cleanup);
  static adamic_string line = ADAMIC_STRING("buffered");
  adamic_write_line(adamic_stdout,&line);
  static char error_buffer[256];
  if (setvbuf(stderr,error_buffer,_IOFBF,sizeof error_buffer)!=0) abort();
  printf("stdio"); fprintf(stderr,"stderr");
  adamic_process_exit_now(atoi(argv[2]));
 } else if (strcmp(mode,"status") == 0) {
  int32_t code;
  log_bool(!adamic_host_exit_code(&code));
  adamic_host_set_exit_code(true,2);
  adamic_host_exit_code(&code); printf("%d\n",code);
  adamic_host_set_exit_code(false,0);
  log_bool(!adamic_host_exit_code(&code));
  puts("continues");
 } else if (strcmp(mode,"natural") == 0) {
  adamic_host_set_exit_code(true,2); puts("natural cleanup");
 } else if (strcmp(mode,"clocks") == 0) {
  double origin = adamic_host_time_origin(), first = adamic_host_performance_now(), previous = first;
  bool ordered = true, fractional = first != floor(first);
  for (int i=0;i<10000;i++) {
   double now = adamic_host_performance_now(); ordered = ordered && now >= previous;
   fractional = fractional || now != floor(now); previous = now;
  }
  log_bool(ordered); log_bool(fractional); log_bool(previous > first);
  log_bool(origin == adamic_host_time_origin());
  log_bool(fabs(origin + adamic_host_performance_now() - adamic_host_date_now()) < 1000);
  double epoch = adamic_host_date_now(); log_bool(epoch == floor(epoch));
 } else if (strcmp(mode,"memory") == 0) {
  adamic_host_memory first, now, final;
  if (!adamic_host_memory_usage(&first)) abort();
  adamic_typed_array *bytes = adamic_typed_array_new(adamic_typed_array_uint8,4*1024*1024);
  memset(bytes->data,42,bytes->length);
  adamic_typed_array *view = adamic_typed_array_subarray(bytes,1,bytes->length,true);
  if (!adamic_host_memory_usage(&now)) abort();
  puts("rss,heapTotal,heapUsed,external,arrayBuffers");
  log_bool(isfinite(now.rss) && isfinite(now.heapTotal) && isfinite(now.heapUsed) && isfinite(now.external) && isfinite(now.arrayBuffers) &&
   now.rss >= 0 && now.heapTotal >= 0 && now.heapUsed >= 0 && now.external >= 0 && now.arrayBuffers >= 0);
  log_bool(now.rss > 0 && now.heapTotal >= now.heapUsed);
  log_bool(now.arrayBuffers-first.arrayBuffers >= bytes->length);
  log_bool(now.external >= now.arrayBuffers);
  adamic_release(bytes);
  log_bool(((unsigned char *)view->data)[0] == 42);
  adamic_release(view);
  if (!adamic_host_memory_usage(&final) || final.arrayBuffers != first.arrayBuffers) return 4;
 } else if (strcmp(mode,"identity") == 0) {
  adamic_array *args = adamic_host_argv(), *exec_args = adamic_host_exec_argv();
  adamic_string *path = adamic_host_exec_path(), *executing = adamic_host_executing_file_path(), *cwd = adamic_host_cwd();
  if (!args || !path || !executing || !cwd) abort();
  log_bool(path->length > 0 && path->bytes[0] == '/');
  char *original = adamic_path_bytes(path);
  char *renamed = malloc(strlen(original) + sizeof ".renamed"); if (!renamed) abort();
  strcpy(renamed, original); strcat(renamed, ".renamed");
  if (rename(original, renamed) != 0) abort();
  adamic_string *again = adamic_host_exec_path();
  if (rename(renamed, original) != 0) abort();
  log_bool(again && again->length == path->length && memcmp(again->bytes,path->bytes,path->length)==0);
  adamic_release(again); free(renamed); free(original);
  adamic_string *first = args->elements[0].reference;
  log_bool(first->length == path->length && memcmp(first->bytes,path->bytes,path->length)==0);
  for (size_t i=3;i<args->length;i++) {
   adamic_string *arg = args->elements[i].reference;
   if (i>3) putchar('|'); printf("%.*s",(int)arg->length,arg->bytes);
  } putchar('\n');
  log_bool(exec_args->length == 0);
  char *file = adamic_path_bytes(executing); char *slash = strrchr(file,'/');
  if (!slash) abort(); size_t prefix = (size_t)(slash-file)+1;
  char *lib = malloc(prefix + sizeof "lib.d.ts"); if (!lib) abort();
  memcpy(lib,file,prefix); memcpy(lib+prefix,"lib.d.ts",sizeof "lib.d.ts");
  log_bool(access(lib,F_OK)==0); free(lib); free(file);
  log_bool(equal(cwd,getenv("HOST_EXPECT_CWD")));
  adamic_release(args); adamic_release(exec_args); adamic_release(path); adamic_release(executing); adamic_release(cwd);
 } else if (strcmp(mode,"cwd") == 0) {
  adamic_string *cached = adamic_host_cwd();
  adamic_string *first = adamic_host_cwd();
  log_bool(first && cached && first->length==cached->length && memcmp(first->bytes,cached->bytes,first->length)==0); adamic_release(first);
  if (chdir(getenv("HOST_CHILD")) != 0) abort();
  adamic_string *now = adamic_host_cwd();
  log_bool(now && cached && (now->length!=cached->length || memcmp(now->bytes,cached->bytes,now->length)!=0));
  log_bool(equal(now,getenv("HOST_CHILD"))); adamic_release(now); adamic_release(cached);
 } else if (strcmp(mode,"environment") == 0) {
  static adamic_string missing = ADAMIC_STRING("HOST_MISSING"), empty = ADAMIC_STRING("HOST_EMPTY"), key = ADAMIC_STRING("HOST_VALUE");
  adamic_string *value = adamic_host_environment(&missing); log_bool(value==NULL); adamic_release(value);
  value = adamic_host_environment(&missing); log_bool(!value || value->length==0); adamic_release(value);
  value = adamic_host_environment(&empty); log_bool(value && value->length==0); adamic_release(value);
  adamic_string *before = adamic_host_environment(&key); print_text(before);
  if (setenv("HOST_VALUE","changed",1)!=0) abort(); // Host-side change; no Adamic env-write API.
  value = adamic_host_environment(&key); print_text(value); print_text(before); adamic_release(value); adamic_release(before);
 } else if (strcmp(mode,"eol") == 0) {
  adamic_string *eol = adamic_host_eol();
  if (equal(eol,"\n")) puts("\"\\n\""); else puts("wrong EOL");
  printf("Version 6.0.3%.*sdiagnostic%.*s",(int)eol->length,eol->bytes,(int)eol->length,eol->bytes);
 } else return 9;
 return adamic_host_exit_status();
}
