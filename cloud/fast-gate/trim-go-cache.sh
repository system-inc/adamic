#!/usr/bin/env bash
# Keeps a gate box's disk from filling with Go's build cache (#93eddxh). Oct 9 05:16Z: Threadripper's disk hit 100% with
# the cache at 537 GB of 1007 GB, every gate there voided on 'No space left on device', and the void storm paused all
# dispatch. Go trims only entries unused for five days, and every candidate adds new actions. Run on the box before a
# gate takes its tree (cloud/fast-gate.sh, cloud/full-gate-main.sh):
#
#   bash cloud/fast-gate/trim-go-cache.sh
#
# Over 80 percent of the cache's filesystem used, it removes the entries unused for 12 hours. That is Go's own trim with
# a shorter window, by modification time, which Go refreshes on use, so gates building in the box's other slots meanwhile
# only miss and rebuild; go clean -cache would pull files out from under them. It prints what it did, and says so when
# the disk is still over 90 percent after the trim.
set -uo pipefail
cache=$(go env GOCACHE 2> /dev/null) || exit 0
[ -d "${cache}" ] || exit 0
used() {
  df -P "${cache}" | awk 'NR == 2 {sub("%", "", $5); print $5}'
}
before=$(used)
[ -n "${before}" ] && [ "${before}" -gt 80 ] || exit 0
find "${cache}" -type f -mmin +720 -delete 2> /dev/null
after=$(used)
echo "trimmed Go's build cache ${cache}: entries unused 12 h removed, disk ${before}% to ${after}% used"
[ "${after}" -le 90 ] || echo "disk still ${after}% used after trimming Go's build cache on $(hostname): the box needs a hand"
exit 0
