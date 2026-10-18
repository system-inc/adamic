#!/bin/sh
case "$1:$2" in
 js:stage3/parser-next/speculation/misfit.a)
  exec /tmp/delta-negatives-slice-adamic js /tmp/delta-negatives-slice5/stage3/parser-next/speculation/misfit.a ;;
 build:stage3/parser-next/speculation/misfit.a)
  exec /tmp/delta-negatives-slice-adamic build /tmp/delta-negatives-slice5/stage3/parser-next/speculation/misfit.a -o "$4" ;;
 *) exec /tmp/delta-negatives-slice-adamic "$@" ;;
esac
