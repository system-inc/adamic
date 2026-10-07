#!/usr/bin/env python3
"""Reduce map.c's numeric hash to initial-bucket populations, not probe counts."""
import collections
import struct

print('Numeric hash from internal/native/runtime/map.c, bucket_count = 2 * key count.')
for count in [1024, 4096, 65536]:
    buckets = collections.Counter()
    for key in range(count):
        bits = struct.unpack('<Q', struct.pack('<d', float(key)))[0]
        bucket = ((bits ^ (bits >> 29)) * 1099511628211) & (count * 2 - 1)
        buckets[bucket] += 1
    print((count, len(buckets), max(buckets.values())))
print('Columns: keys, occupied initial buckets, maximum population.')
