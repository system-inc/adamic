# #r11wppy's and #96pc420's reading: builds tsc's adapted tree at this commit's parent (main 03c140ec, the scan's)
# and prints each listed site with five lines either side. Run after environment.sh, sealed-job.sh,
# sealed_typescript_source and sealed_npm stage3/api. Never merged.
set -u
D=$(cd "$(dirname "$0")" && pwd)
A=$TMPDIR/adapted
bash stage3/apply.sh "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -5 "$TMPDIR/apply.log"; exit 3; }
for list in promise-sites never-sites; do
  echo "######## $list"
  while IFS=$'\t' read -r site words; do
    file=${site%%:*}; rest=${site#*:}; line=${rest%%:*}
    echo "==== $site | $words"
    start=$(( line > 5 ? line - 5 : 1 ))
    awk -v s="$start" -v e="$(( line + 5 ))" -v l="$line" 'NR>=s && NR<=e { printf "%s%6d  %s\n", (NR==l ? ">" : " "), NR, $0 }' "$A/$file" | cut -c1-200
  done < "$D/$list.tsv"
done
echo "VERDICT OK"
