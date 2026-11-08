#!/bin/bash
# Makes every later shell source setup.sh's env.sh. Arguments: the env.sh path, and the home directory to write in.
#
# An interactive shell reads ~/.bashrc. A login shell reads ~/.bash_profile, or ~/.profile when there is none, and
# Debian's ~/.profile includes ~/.bashrc, whose first lines return at once in a shell that isn't interactive. Codex
# runs every agent command as `bash -lc`, a login shell that isn't interactive, so a line only at the end of
# ~/.bashrc never ran there: the agent found the image's /usr/bin/go, a board game, and no clang (#ybkwr91, Codex
# session 01a1bc4d, Nov 8). So the line goes in ~/.bashrc and in the file a login shell reads, once in each.
set -euo pipefail
environment=$1 home=$2
line="source $environment"
login="$home/.profile"
[ -e "$home/.bash_profile" ] && login="$home/.bash_profile"
for profile in "$home/.bashrc" "$login"; do
	grep -qsxF "$line" "$profile" || echo "$line" >> "$profile"
done
