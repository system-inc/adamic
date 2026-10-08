#!/usr/bin/env bash
# Runs a long job on the gate box inside its thread budget, so it can't slow a landing check:
#
#   bash cloud/box-run.sh idle <command...>
#
# The fast gate's two slots each own half the CPUs (cloud/fast-gate.sh). Everything else, the full
# gate on main, an area merge, a measurement, runs idle-scheduled across all CPUs: SCHED_IDLE, the
# idle I/O class and nice 19, so it gets only the cycles and disk the slots leave and finishes later
# rather than making a landing check's timing depend on it.
set -euo pipefail
class=${1:?usage: box-run.sh idle <command...>}
shift
case ${class} in
  idle) exec chrt --idle 0 ionice -c3 nice -n 19 "$@" ;;
  *) echo "unknown class ${class}; the fast gate's slots are taken only through cloud/fast-gate.sh" >&2; exit 2 ;;
esac
