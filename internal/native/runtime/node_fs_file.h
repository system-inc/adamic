// fs file host entry points. Inputs are borrowed; pointer results are owned.
#ifndef ADAMIC_NODE_FS_FILE_H
#define ADAMIC_NODE_FS_FILE_H
adamic_string *adamic_fs_file_uncaught_text(const adamic_object *);
void adamic_fs_file_error_classes(const adamic_class *, const adamic_class *, const adamic_class *);
adamic_string *adamic_fs_file_read_file(const adamic_string *, const adamic_string *);
adamic_array *adamic_fs_file_read_buffer(const adamic_string *, const adamic_string *);
adamic_array *adamic_fs_file_read_buffer_fd(double, const adamic_string *);
adamic_string *adamic_fs_file_read_fd(double, const adamic_string *);
double adamic_fs_file_open(const adamic_string *, const adamic_string *, double);
double adamic_fs_file_read_sync(double, adamic_array *, double, double, double);
double adamic_fs_file_write(double, const adamic_string *, double);
double adamic_fs_file_close(double);
double adamic_fs_file_write_buffer(const adamic_string *, const adamic_array *, const adamic_string *, double, bool);
double adamic_fs_file_write_buffer_fd(double, const adamic_array *, const adamic_string *, double, bool);
double adamic_fs_file_write_file(const adamic_string *, const adamic_string *, const adamic_string *, double, bool);
double adamic_fs_file_write_fd(double, const adamic_string *, const adamic_string *, double, bool);
bool adamic_fs_file_exists(const adamic_string *);
adamic_object *adamic_fs_file_stat(const adamic_string *, bool);
adamic_string *adamic_fs_file_mkdir(const adamic_string *, bool, double);
double adamic_fs_file_unlink(const adamic_string *);
double adamic_fs_file_utimes(const adamic_string *, double, double);
double adamic_fs_file_utimes_dates(const adamic_string *, const adamic_object *, const adamic_object *);
double adamic_fs_file_utimes_atime_date(const adamic_string *, const adamic_object *, double);
double adamic_fs_file_utimes_mtime_date(const adamic_string *, double, const adamic_object *);
bool adamic_fs_file_is_file(const adamic_object *);
bool adamic_fs_file_is_directory(const adamic_object *);
bool adamic_fs_file_is_symbolic_link(const adamic_object *);
adamic_object *adamic_fs_file_date_new(double);
double adamic_fs_file_date_time(const adamic_object *);
// Raw bytes seam for the Buffer unit. The caller frees bytes, closes only fds it
// opened, and raises the saved errno after any cleanup. No fake Buffer type.
int adamic_fs_file_read_bytes(int, unsigned char **, size_t *);
#endif
