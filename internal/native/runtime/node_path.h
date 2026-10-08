// POSIX operations selected by the host census.
adamic_string *adamic_node_path_resolve(size_t count,
										adamic_string *const paths[]);
adamic_string *adamic_node_path_join(size_t count,
									 adamic_string *const paths[]);
adamic_string *adamic_node_path_dirname(const adamic_string *path);

adamic_string *adamic_node_path_relative(const adamic_string *from,
										 const adamic_string *to);

adamic_string *adamic_node_path_basename(const adamic_string *path, const adamic_string *suffix);

adamic_string *adamic_node_path_normalize(const adamic_string *path);
adamic_string *adamic_node_path_extname(const adamic_string *path);
bool adamic_node_path_isAbsolute(const adamic_string *path);
