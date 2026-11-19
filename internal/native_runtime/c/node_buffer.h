#ifndef ADAMIC_NODE_BUFFER_H
#define ADAMIC_NODE_BUFFER_H
// Encodings: UTF-8 0, UTF-16LE 1, base64 2, hex 3. A Buffer is a Uint8Array,
// one byte an element, as on Node. Inputs are borrowed; Buffer results are owned.
adamic_typed_array *adamic_node_buffer_from(const adamic_string *text, int encoding);
adamic_typed_array *adamic_node_buffer_copy(const adamic_typed_array *source);
adamic_string *adamic_node_buffer_string(const adamic_typed_array *buffer,
										 double start, double end,
										 int encoding);
double adamic_node_buffer_set(adamic_typed_array *buffer, double index, double value);
#endif
