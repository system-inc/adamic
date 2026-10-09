#include "tsgo.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char **argv) {
 assert(argc == 4);
 size_t length = strlen(argv[1]);
 char *config = malloc(length);
 assert(config != NULL);
 memcpy(config, argv[1], length);
 tsgo_view view = {config, length};
 if (strcmp(argv[3], "length") == 0) view.length++;
 tsgo_handle handle = 0;
 tsgo_buffer error = {0};
 assert(tsgo_create(view, NULL, 0, &handle, &error) == TSGO_OK);
 assert(handle != 0 && error.data == NULL);
 free(config);
 tsgo_view file = {argv[2], strlen(argv[2])};
 tsgo_result answer = {0};
 for (size_t index = 0; index < 100; index++) {
  assert(tsgo_query(handle, file, 14, &answer, &error) == TSGO_OK);
  assert(answer.kind != 0 && answer.type.length != 0);
  tsgo_result_free(&answer);
  assert(answer.type.data == NULL && answer.type.length == 0);
 }
 assert(tsgo_query(handle, file, UINT64_MAX, &answer, &error) == TSGO_CHECKER);
 assert(error.length != 0);
 tsgo_buffer_free(&error);
 assert(tsgo_query(handle, file, 14, &answer, &error) == TSGO_OK);
 assert(tsgo_release(handle, &error) == TSGO_OK);
 /* Output buffers must survive their program. */
 assert(fwrite(answer.type.data, 1, answer.type.length, stdout) == answer.type.length);
 tsgo_result_free(&answer);
 assert(tsgo_query(handle, file, 14, &answer, &error) == TSGO_HANDLE);
 tsgo_buffer_free(&error);
 assert(tsgo_release(handle, &error) == TSGO_HANDLE);
 tsgo_buffer_free(&error);
 tsgo_handle other = 0;
 tsgo_view correct = {argv[1], strlen(argv[1])};
 assert(tsgo_create(correct, NULL, 0, &other, &error) == TSGO_OK);
 assert(other != handle);
 assert(tsgo_release(other, &error) == TSGO_OK);
 assert(tsgo_query(0, file, 14, &answer, &error) == TSGO_HANDLE);
 tsgo_buffer_free(&error);
 assert(tsgo_create((tsgo_view){NULL, 1}, NULL, 0, &other, &error) == TSGO_ARGUMENT);
 assert(tsgo_query(0, (tsgo_view){NULL, 1}, 0, &answer, &error) == TSGO_ARGUMENT);
 return 0;
}
