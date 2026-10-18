#!/bin/sh
if [ -n "$ADAMIC_RSS_DIR" ]; then
 record=$(mktemp "$ADAMIC_RSS_DIR/clang-XXXXXX.rss")
 exec /tmp/v3-json-measure/rss "$record" /workspace/adamic-tools/llvm/bin/clang "$@"
fi
exec /workspace/adamic-tools/llvm/bin/clang "$@"
