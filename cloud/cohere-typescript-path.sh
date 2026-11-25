# cloud/cohere-typescript-path.sh: cohere_typescript_path, the one reader of where cohere keeps its TypeScript submodule,
# for every script under cloud/ that lays or names it. Sourced, never run, and it defines this one function and nothing
# else, so a script with its own environment (an image builder under set -euo pipefail, a meter job, a probe) sources it
# without taking sandboxed-job.sh's steps; sandboxed-job.sh sources it too, so a sealed job has it after the preamble.
#
#   . cloud/cohere-typescript-path.sh && typescriptPath=$(cohere_typescript_path [<adamic root>])
#
# cohere_typescript_path [root]: the path of cohere's TypeScript submodule relative to cohere's root, on stdout
# (TypeScript at cohere's pin as of Nov 25, a path under libraries/typescript once cohere moves it), read from
# <root>/cohere/.gitmodules as the tree holds it, so the answer is always the pinned commit's. The root is adamic's, and
# . when none is named. The entry is the one named TypeScript, exactly as git spells it, since cohere's move keeps that
# name and changes only the path. Its url decides nothing, so a mirror or a fork in the url changes nothing, and an
# entry renamed away from TypeScript is refused rather than found by what it points at. It never falls back to a
# default: no .gitmodules, no entry named TypeScript with a path, more than one, or a path that is absolute, holds a
# `..` or `.` part, or isn't a directory in the tree prints one FAIL line on stderr naming what it read, and returns 1.
# cloud/cohere-typescript-path-test.sh holds it to each of those.
cohere_typescript_path() {
	local root=${1:-.} gitmodules key value name url entries="" paths=() path
	gitmodules="$root/cohere/.gitmodules"
	if [ ! -f "$gitmodules" ]; then
		echo "FAIL cohere_typescript_path: $gitmodules isn't there, so cohere isn't laid and its TypeScript submodule's path can't be read" >&2
		return 1
	fi
	# -z gives each entry as its key, a newline and its value, ended by a NUL, so a name or path with a space reads whole.
	# --file reads that one file alone, never a global, system or repository config. Two sections both named TypeScript
	# give two path values under the one name, so each is counted and two refuse.
	while IFS=$'\n' read -r -d '' key value; do
		name=${key#submodule.}
		name=${name%.path}
		# The url is read only to name the entry in a refusal, so a renamed entry says what it points at.
		url=$(git config --no-includes --file "$gitmodules" --get-all "submodule.$name.url" 2> /dev/null | head -n 1)
		entries="$entries${entries:+, }$name at $value${url:+ from $url}"
		if [ "$name" = TypeScript ]; then
			paths+=("$value")
		fi
	done < <(git config --no-includes --file "$gitmodules" -z --get-regexp '^submodule\..*\.path$' 2> /dev/null)
	if [ ${#paths[@]} -ne 1 ]; then
		echo "FAIL cohere_typescript_path: $gitmodules has ${#paths[@]} submodule paths named TypeScript, not one (it read: ${entries:-no submodule path})" >&2
		return 1
	fi
	path=${paths[0]}
	case "/$path/" in
		//* | */../* | */./* | *//*)
			echo "FAIL cohere_typescript_path: $gitmodules gives the TypeScript submodule the path '$path', which is no relative path inside cohere" >&2
			return 1
			;;
	esac
	if [ ! -d "$root/cohere/$path" ] || [ -L "$root/cohere/$path" ]; then
		echo "FAIL cohere_typescript_path: $gitmodules gives the TypeScript submodule the path $path, and $root/cohere/$path is no directory of the tree's" >&2
		return 1
	fi
	printf '%s\n' "$path"
}
