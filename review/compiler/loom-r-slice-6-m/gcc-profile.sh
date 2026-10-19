#!/bin/sh
if [ "$1" = --version ]; then
 echo 'GCC 14 profile with LLVM 20.1.8 sanitizer allocator interface headers'
 exec /usr/bin/gcc "$@"
fi
exec /usr/bin/gcc -isystem /tmp/adamic-gate/loom-r-slice-6-m/gcc-headers "$@"
