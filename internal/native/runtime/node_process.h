#ifndef ADAMIC_NODE_PROCESS_H
#define ADAMIC_NODE_PROCESS_H
bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text);
void adamic_node_process_start(int count, char **values);
adamic_string *adamic_node_cwd(void);
void adamic_node_chdir(const adamic_string *directory);
adamic_string *adamic_node_platform(void);
adamic_string *adamic_node_eol(void);
adamic_string *adamic_node_tmpdir(void);
void adamic_node_error(adamic_string *name, adamic_string *message, adamic_string *code);
bool adamic_node_next_tick_feature(void);
double adamic_node_pid(void);
adamic_array *adamic_node_argv(void);
adamic_array *adamic_node_exec_argv(void);
adamic_maybe_number adamic_node_columns(void);
adamic_object *adamic_node_stdout_handle(void);
bool adamic_node_stdout_write(const adamic_string *text);
adamic_object *adamic_node_memory_usage(void);
double adamic_node_performance_now(void);
double adamic_node_time_origin(void);
adamic_object *adamic_node_mark(const adamic_string *name);
adamic_object *adamic_node_measure(const adamic_string *name, const adamic_string *start, const adamic_string *end);
void adamic_node_clear_marks(const adamic_string *name);
void adamic_node_clear_measures(const adamic_string *name);
void adamic_node_environment_set(const adamic_string *name, const adamic_string *value);
bool adamic_node_environment_delete(const adamic_string *name);
adamic_object *adamic_node_performance(void);
adamic_value adamic_node_performance_invoke(adamic_closure *closure, adamic_value *args, size_t count, bool discard);
#endif
