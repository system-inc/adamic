# Shared shell child deadlines. GNU timeout gives each invocation a process group.
# Cold setup observed 114s; 10m leaves room for cold toolchain work and downloads.
# An optional shorter cap supports individual child budgets and hang proofs.
# Aggregate supervisors above 600s retain their limit so they can reap timed-out children.
exec {ADAMIC_BOUNDED_REPORT}>&2
bounded() {
 local limit=$1 status=0
 shift
 if [ "$limit" -le 600 ] && [ -n "${ADAMIC_CHILD_DEADLINE:-}" ]; then
  local cap=$ADAMIC_CHILD_DEADLINE whole
  if [[ ! "$cap" =~ ^[0-9]+([.][0-9]+)?$ ]] || [[ "$cap" =~ ^0+([.]0+)?$ ]]; then
   echo "ADAMIC_CHILD_DEADLINE must be positive seconds" >&"$ADAMIC_BOUNDED_REPORT"
   return 2
  fi
  whole=${cap%%.*}
  if (( 10#$whole < limit )); then limit=$cap; fi
 fi
 # Keep the timeout leader alive after TERM. GNU timeout otherwise cancels
 # --kill-after when that leader exits, orphaning TERM-ignoring grandchildren.
 # Explicit stdin inheritance preserves tar and other pipeline consumers.
 timeout --kill-after=1s "$limit" bash -c '
  limit=$1 report=$2
  shift 2
  expired=0
  child_name=$1
  on_deadline() { expired=1; printf "child %s: deadline %s exceeded; killing process group\n" "$child_name" "$limit" >&"$report"; }
  trap on_deadline TERM
  command "$@" <&0 & child=$!
  wait "$child"; status=$?
  if [ "$expired" = 1 ]; then
   while :; do sleep 1; done
  fi
  exit "$status"
 ' bounded-child "$limit" "$ADAMIC_BOUNDED_REPORT" "$@" || status=$?
 return "$status"
}
