#ifndef ADAMIC_NODE_CWD_WASI_H
#define ADAMIC_NODE_CWD_WASI_H
#ifdef ADAMIC_TARGET_WASI
// Initialize the command host's cwd on its first filesystem/cwd observation.
// Shared across adapters; later chdir calls must never restore the startup cwd.
void adamic_node_wasi_cwd(void);
#endif
#endif
