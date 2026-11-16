#!/bin/bash
# #re2xhwa: does each holdout port build natively, and print what the game's original printed at the
# game's size? Output only, never a time. Hashes are the originals' on Node (bench/holdout/README.md).
# Never lands.
set -u
T=$(mktemp -d); H=bench/holdout; failed=0
go build -buildvcs=false -o $T/adamic ./cmd/adamic || { echo "FAIL build adamic"; exit 1; }
for p in fasta fannkuchredux knucleotide mandelbrot regexredux revcomp; do
  if $T/adamic build $H/$p.a -o $T/$p > $T/$p.build 2>&1; then echo "PASS built $p"; else echo "FAIL build $p: $(grep -m3 -E 'NotYet|refus|error|can.t' $T/$p.build | tr '\n' ' ' | cut -c1-400)"; failed=1; fi
done
check() { # name want got
  if [ "$2" = "$3" ]; then echo "PASS $1 matches the original: $3"; else echo "FAIL $1: sha256 $3, the original's $2"; failed=1; fi
}
if [ -x $T/fasta ]; then
  $T/fasta 25000000 > $T/in25.txt; check "fasta 25,000,000" 3fcf4f78104c8a65ef210fe1d469f4e473456c791225f2f1f9114f4986aa09fa "$(sha256sum < $T/in25.txt | cut -d' ' -f1)"
  $T/fasta 5000000 > $T/in5.txt; check "fasta 5,000,000" 97197f5957a12f8a859ba7edff6d97994daa18afacd77db21fa68c6f2e447e38 "$(sha256sum < $T/in5.txt | cut -d' ' -f1)"
fi
[ -x $T/knucleotide ] && [ -f $T/in25.txt ] && check knucleotide f8c042fe770dda8f870b967605ea40ca2e353f2bdf92e89a327209dd31d445c7 "$($T/knucleotide $T/in25.txt | sha256sum | cut -d' ' -f1)"
[ -x $T/regexredux ] && [ -f $T/in5.txt ] && check regexredux a37e6da1d43b332d5c0f4fc25b76090e9095d4c97c25f1a7186385e87706e4af "$($T/regexredux $T/in5.txt | sha256sum | cut -d' ' -f1)"
[ -x $T/revcomp ] && [ -f $T/in25.txt ] && check revcomp e5730d4fd55c464f276e5c6062c80f46e0a735281de15d8c957e2731a9baf70e "$($T/revcomp $T/in25.txt | sha256sum | cut -d' ' -f1)"
[ -x $T/fannkuchredux ] && echo "fannkuchredux prints: $($T/fannkuchredux | tr '\n' '|') (the game's n = 12 answer is 3968050|Pfannkuchen(12) = 65|)"
[ -x $T/mandelbrot ] && echo "mandelbrot sha256 $($T/mandelbrot | sha256sum | cut -d' ' -f1); the original's at 16000 isn't recorded, so this is the first reading"
exit $failed
