adamic_array *adamic_node_fs_readdir(const adamic_string *path,
									 const adamic_object *options);
bool adamic_node_fs_dirent_is(const adamic_object *entry, const char *method);
adamic_string *adamic_node_fs_realpath(const adamic_string *path, bool native);
void adamic_node_fs_raise(const adamic_string *path, int error,
						  const char *operation);
adamic_object *adamic_real_path_node(const adamic_string *path);
