#include <dirent.h>
#include <errno.h>
#include <stdio.h>
int main(int argc, char **argv) {
 if (argc != 2) return 2;
 DIR *directory = opendir(argv[1]);
 if (!directory) { printf("open errno=%d\n", errno); return 1; }
 errno = 0;
 while (readdir(directory)) {}
 int error = errno;
 closedir(directory);
 printf("readdir errno=%d\n", error);
 return error ? 1 : 0;
}
