#ifndef ADAMIC_NODE_BUFFER_H
#define ADAMIC_NODE_BUFFER_H
// Encodings: UTF-8 0, UTF-16LE 1, base64 2, hex 3, and 'ascii' 4, which only
// adamic_node_buffer_string takes. A Buffer is a Uint8Array,
// one byte an element, as on Node. Inputs are borrowed; Buffer results are owned.
adamic_typed_array *adamic_node_buffer_from(const adamic_string *text, int encoding);
adamic_typed_array *adamic_node_buffer_copy(const adamic_typed_array *source);
adamic_string *adamic_node_buffer_string(const adamic_typed_array *buffer,
										 double start, double end,
										 int encoding);
double adamic_node_buffer_set(adamic_typed_array *buffer, double index, double value);
// Buffer.allocUnsafe(size): an owned Buffer, or NULL with Node's coded
// RangeError pending for a size out of range or one that can't be allocated.
adamic_typed_array *adamic_node_buffer_alloc_unsafe(double size);
// Buffer.indexOf(value, byteOffset): the first index at or past the offset
// where value's bytes are, or -1. A string, a literal free of lone
// surrogates, is searched as its UTF-8, and a number as one byte, masked to
// 8 bits.
// has_offset false is an absent or undefined offset. Neither throws.
double adamic_node_buffer_index_of_string(const adamic_typed_array *buffer,
										  const adamic_string *needle,
										  double offset, bool has_offset);
double adamic_node_buffer_index_of_number(const adamic_typed_array *buffer,
										  double needle, double offset,
										  bool has_offset);
#endif
