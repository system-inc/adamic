#ifndef ADAMIC_NODE_CRYPTO_H
#define ADAMIC_NODE_CRYPTO_H
adamic_object *adamic_node_hash_new(void);
adamic_object *adamic_node_hash_update(adamic_object *hash,
									   const adamic_string *text, int encoding);
adamic_string *adamic_node_hash_digest(adamic_object *hash);
#endif
