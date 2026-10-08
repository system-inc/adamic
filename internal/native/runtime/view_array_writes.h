#ifndef ADAMIC_VIEW_ARRAY_WRITES_H
#define ADAMIC_VIEW_ARRAY_WRITES_H
struct adamic_object;
struct adamic_array;
typedef struct adamic_array_write_pair { unsigned int source, target; } adamic_array_write_pair;
void adamic_view_array_object_certificate(struct adamic_object *object, unsigned int contract);
void adamic_view_array_element_certificate(struct adamic_array *array, unsigned int contract);
void adamic_view_array_reference_write(const struct adamic_array *array, const struct adamic_object *value, const adamic_array_write_pair *pairs, size_t count, const char *const *names, size_t name_count);
#endif
