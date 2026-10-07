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
 timeout --kill-after=1s "$limit" "$@" || status=$?
 if [ "$status" = 124 ] || [ "$status" = 137 ]; then
  echo "child $1: deadline ${limit} exceeded; process group killed" >&"$ADAMIC_BOUNDED_REPORT"
 fi
 return "$status"
}
