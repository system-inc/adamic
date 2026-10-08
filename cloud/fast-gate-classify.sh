# A tip's class for scheduling only (coverage is the gate's business): big for an area, a stage3/
# change, or more than two touched packages (directories of changed Go files, a testdata path counting
# as its package); small otherwise. A big gate runs in the area slot (24 CPUs), a small one in either
# small slot (12 each). Sourced by cloud/fast-gate-watch.sh, which classes each tip when it queues it,
# and by cloud/fast-gate.sh, which classes a gate run without --class (integration's direct landing
# gates) the same way, so a two-file fix-forward never waits for the area slot. Needs ${here}.
classify() {
  local branch=$1 sha=$2 count changed
  [[ ${branch} == area/* ]] && { echo B; return; }
  git -C "${here}" fetch -q origin "${sha}" 2>/dev/null || { echo B; return; }
  changed=$(git -C "${here}" diff --name-only "$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)...${sha}" 2>/dev/null)
  # A stage3/ change runs the stage 3 lane (about 10 minutes): big, whatever else it touches.
  echo "${changed}" | grep -q '^stage3/' && { echo B; return; }
  count=$(echo "${changed}" | awk '/\/testdata\// {sub("/testdata/.*", ""); print; next} /\.go$/ {sub("/[^/]*$", ""); print}' | sort -u | wc -l)
  [ "${count}" -gt 2 ] && echo B || echo S
}
