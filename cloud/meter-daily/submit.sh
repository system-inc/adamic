#!/usr/bin/env bash
# cloud/meter-daily/submit.sh hands a finished daily meter run to Loom's Queue, as ahra. meter-daily.service starts
# it (OnSuccess=) after run.sh exits 0; by hand it reports the last run and does the same.
#
#   bash cloud/meter-daily/submit.sh [--dry-run]
#
# It reads run.sh's status file, and only a Succeeded run goes further; a red run commits nothing and its logs stay
# on the box. The run directory, read and never written, is copied to staging, where each .log is gzipped to .log.gz
# (the shape of the hand-committed 20261020T014828Z.POASDN; the root .gitignore ignores raw .log). What git's ignore
# rules then keep is committed alone on a fresh branch cut from origin/main's tip, nothing force-added: so
# stage3/meter/.gitignore's census.jsonl.gz and latent.jsonl.gz stay out until that file says otherwise. It is
# pushed with a deploy key scoped to adamic, and submitted for system_adamic_compiler_correctness with a
# submit-only token: the same POST /changes that `ahra loom submit <branch> --owner <owner> --yes` makes with
# --base off, whose body carries the merge base of the tip and origin/main. A write is never retried.
#
# Both secrets are paths named in /etc/adamic-meter/submit.env (ADAMIC_METER_SUBMIT_ENV overrides it):
#
#   ADAMIC_METER_DEPLOY_KEY=/etc/adamic-meter/deploy-key       an ssh key with write access to adamic only
#   ADAMIC_METER_SUBMIT_TOKEN=/etc/adamic-meter/submit-token   a Queue token that can submit for the owner
#
# Until both are there and readable it exits 3, "not configured", having pushed and submitted nothing.
# --dry-run builds the commit and the request body and prints them, still pushing and posting nothing, so the
# commit can be read before the secrets exist. A failed meter run exits 1 here with its step and log; there is
# no route a box can post a task comment through yet, so that report is this script's output and the journal.
#
# The bare clone, the staged copy of each run and what was pushed and submitted stay in $STATE_DIRECTORY (from
# the unit, else /var/lib/adamic-meter-submit), so a rerun after a refusal submits the branch it already pushed
# rather than pushing another, and a run that was submitted is never submitted twice.
set -euo pipefail

[ "$(uname)" = Linux ] || { echo "meter-daily-submit: Linux build boxes only, never a Mac" >&2; exit 2; }

isDryRun=false
for argument in "$@"; do
	case "$argument" in
		--dry-run) isDryRun=true ;;
		*) echo "usage: bash cloud/meter-daily/submit.sh [--dry-run]" >&2; exit 2 ;;
	esac
done

meterHome=${ADAMIC_METER_HOME:-/var/lib/adamic-meter}
state=${STATE_DIRECTORY:-/var/lib/adamic-meter-submit}
environment=${ADAMIC_METER_SUBMIT_ENV:-/etc/adamic-meter/submit.env}
owner=system_adamic_compiler_correctness
queue=https://loom.system.inc
origin=https://github.com/system-inc/adamic.git
pushOrigin=git@github.com:system-inc/adamic.git
status=$meterHome/status

say() { echo "meter-daily-submit: $*"; }
# field <key>: one line of run.sh's status file
field() { sed -n "s/^$1=//p" "$status" | head -n 1; }

# The last meter run, reported first whatever happens next.
outcome=None
if [ -r "$status" ]; then
	outcome=$(field outcome)
	say "last meter run: $outcome at step $(field step), main $(field sha), started $(field started), log $(field log)"
	[ "$outcome" = Succeeded ] || [ -z "$(field detail)" ] || say "  $(field detail)"
else
	say "no meter run recorded ($status is absent or unreadable)"
fi

# The unit reads the env file itself (EnvironmentFile=-); a hand run reads the same two keys from it here.
if [ -r "$environment" ]; then
	for key in ADAMIC_METER_DEPLOY_KEY ADAMIC_METER_SUBMIT_TOKEN; do
		[ -n "${!key:-}" ] || export "$key=$(sed -n "s/^$key=//p" "$environment" | tail -n 1 | sed 's/^"\(.*\)"$/\1/')"
	done
fi
problems=""
[ -r "$environment" ] || problems="$problems; $environment is absent or unreadable"
for key in ADAMIC_METER_DEPLOY_KEY ADAMIC_METER_SUBMIT_TOKEN; do
	if [ -z "${!key:-}" ]; then
		problems="$problems; $key is unset"
	elif [ ! -r "${!key}" ]; then
		problems="$problems; $key names ${!key}, which is absent or unreadable"
	fi
done
if [ -n "$problems" ] && ! "$isDryRun"; then
	say "not configured: ${problems#; }. Nothing pushed, nothing submitted."
	exit 3
fi

[ "$outcome" = Succeeded ] || { say "no finished run to submit. Nothing pushed, nothing submitted."; exit 1; }
run=$(field run)
measured=$(field sha)
stamp=$(basename "$run")
[ -f "$run/report.md" ] && [ -f "$run/report.json" ] || { say "$run holds no report. Nothing pushed, nothing submitted."; exit 1; }
mkdir -p "$state/staging" "$state/pushed" "$state/submitted"
if [ -f "$state/submitted/$stamp" ]; then
	say "$stamp was submitted already: $(cat "$state/submitted/$stamp")"
	exit 0
fi

export GIT_DIR=$state/adamic.git
# Main's tip is all a run's commit needs: its parent, and the trees git add and the diff below read. Depth 1 and
# no blobs keep the clone a few MB; a blob it lacks (an ignore file) is fetched when it is read.
[ -d "$GIT_DIR" ] || git clone --quiet --bare --depth 1 --filter=blob:none --single-branch --branch main "$origin" "$GIT_DIR"
git fetch --quiet origin main:refs/remotes/origin/main

branch=meter/daily-$stamp
if [ -f "$state/pushed/$stamp" ]; then
	# Pushed by an earlier attempt whose submit failed: submit that very branch, never a second one.
	tip=$(sed -n 's/^tip=//p' "$state/pushed/$stamp")
	say "$branch was pushed already at $tip; submitting it"
else
	base=$(git rev-parse --verify 'refs/remotes/origin/main^{commit}')
	# A work tree holding only the run and the ignore files on its path, as origin/main has them, so git add keeps
	# exactly what git's rules keep. Copied whole first, then moved, so a rerun never adds a half-copied run.
	staging=$state/staging/$stamp
	target=$staging/stage3/meter/runs/$stamp
	if [ ! -d "$target" ]; then
		mkdir -p "$staging/stage3/meter/runs"
		for prefix in "" stage3/ stage3/meter/ stage3/meter/runs/; do
			if git cat-file -e "$base:${prefix}.gitignore" 2> /dev/null; then
				git show "$base:${prefix}.gitignore" > "$staging/${prefix}.gitignore"
			fi
		done
		cp -R "$run" "$target.partial"
		# -n keeps the name and time out of the header, so the same log always gzips to the same bytes.
		find "$target.partial" -type f -name '*.log' -exec gzip -n -- {} +
		mv "$target.partial" "$target"
	fi
	index=$state/staging/$stamp.index
	GIT_INDEX_FILE=$index git read-tree "$base"
	GIT_INDEX_FILE=$index git --work-tree="$staging" add -- "stage3/meter/runs/$stamp"
	tree=$(GIT_INDEX_FILE=$index git write-tree)
	# The commit holds this run and nothing else, its report among it.
	changed=$(git diff --name-only --no-renames "$base" "$tree")
	[ -n "$changed" ] || { say "the run adds nothing to origin/main"; exit 1; }
	stray=$(printf '%s\n' "$changed" | grep -v "^stage3/meter/runs/$stamp/" || true)
	[ -z "$stray" ] || { say "the commit would touch paths outside the run: $stray"; exit 1; }
	printf '%s\n' "$changed" | grep -qx "stage3/meter/runs/$stamp/report.md" || { say "the commit leaves out report.md"; exit 1; }
	message=$state/staging/$stamp.message
	{
		printf 'stage3/meter: the daily run %s on main %s\n\n' "$stamp" "${measured:0:12}"
		printf 'Measured by meter-daily.service on %s, as adamicmeter, from main %s.\n' "$(hostname)" "$measured"
		printf '%s\n\nTask: #badrnrf\n' "$(head -n 2 "$run/report.md")"
	} > "$message"
	tip=$(git -c user.name="${ADAMIC_METER_AUTHOR_NAME:-meter-daily on $(hostname)}" -c user.email="${ADAMIC_METER_AUTHOR_EMAIL:-ahra@ahra.ai}" \
		commit-tree "$tree" -p "$base" -F "$message")
	say "committed $tip on $base: $(printf '%s\n' "$changed" | wc -l | tr -d ' ') files under stage3/meter/runs/$stamp"
	if "$isDryRun"; then
		git show --stat --format='%H%n%an <%ae>%n%n%B' "$tip"
	else
		# GitHub's host keys from its own API over TLS, so the first push trusts no key it was only shown.
		if [ ! -s "$state/known_hosts" ]; then
			curl -fsS --max-time 30 https://api.github.com/meta |
				python3 -I -c 'import json, sys; print("\n".join("github.com " + key for key in json.load(sys.stdin)["ssh_keys"]))' > "$state/known_hosts.writing"
			mv "$state/known_hosts.writing" "$state/known_hosts"
		fi
		GIT_SSH_COMMAND="ssh -i $ADAMIC_METER_DEPLOY_KEY -o IdentitiesOnly=yes -o BatchMode=yes -o StrictHostKeyChecking=yes -o UserKnownHostsFile=$state/known_hosts" \
			git push --quiet "$pushOrigin" "$tip:refs/heads/$branch"
		printf 'branch=%s\ntip=%s\nbase=%s\n' "$branch" "$tip" "$base" > "$state/pushed/$stamp"
		say "pushed $branch at $tip"
	fi
fi

# What `ahra loom submit` posts with --base off: the merge base with origin/main as this clone fetched it, and the
# paths base..tip touches, both sides of a rename listed as the Queue's own diff lists them.
mergeBase=$(git merge-base "$tip" refs/remotes/origin/main)
body=$state/staging/$stamp.body.json
git diff --name-only --no-renames "$mergeBase..$tip" |
	python3 -I -c 'import json, sys
print(json.dumps({"sha": sys.argv[1], "base": sys.argv[2], "owner": sys.argv[3], "paths": [line for line in sys.stdin.read().split("\n") if line]}))' \
		"$tip" "$mergeBase" "$owner" > "$body"
if "$isDryRun"; then
	say "dry run: would POST $queue/changes with Authorization: Bearer <submit token> and this body:"
	python3 -I -m json.tool "$body"
	say "dry run: nothing pushed, nothing submitted${problems:+ (and not configured: ${problems#; })}"
	exit 0
fi

# The token is a one-day token another seat renews, so it is read from its file fresh on every run and never
# kept. It rides a header file, removed as soon as curl has sent it, never the command line any process can read.
headers=$state/staging/$stamp.headers
token=$(tr -d '[:space:]' < "$ADAMIC_METER_SUBMIT_TOKEN")
[ -n "$token" ] || { say "$ADAMIC_METER_SUBMIT_TOKEN is empty. Pushed $branch, submitted nothing."; exit 1; }
(umask 077 && printf 'Authorization: Bearer %s\nContent-Type: application/json\n' "$token" > "$headers")
response=$state/staging/$stamp.response.json
code=$(curl -sS --max-time 20 -o "$response" -w '%{http_code}' -X POST -H "@$headers" --data-binary "@$body" "$queue/changes" || echo 000)
rm -f "${headers:?}"
case $code in
	2??)
		change=$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1])).get("change", ""))' "$response" 2> /dev/null || true)
		printf 'change=%s\nbranch=%s\ntip=%s\nhttp=%s\n' "$change" "$branch" "$tip" "$code" > "$state/submitted/$stamp"
		if [ -n "$change" ]; then
			say "submitted $branch as $change; follow it: ahra loom branch $change"
		else
			say "submitted $branch, but the Queue's answer named no change; find it on ahra loom board rather than submitting again"
		fi
		;;
	*)
		reason=$(python3 -I -c 'import json, sys; answer = json.load(open(sys.argv[1])); print(answer.get("reason") or answer.get("error") or "")' "$response" 2> /dev/null || head -c 300 "$response" 2> /dev/null || true)
		say "the Queue refused $branch: HTTP $code ${reason}. Pushed, not submitted; a rerun submits the same branch."
		exit 1
		;;
esac
