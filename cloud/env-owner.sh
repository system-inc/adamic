# Sourced by setup.sh and write-env.sh. A box's env.sh that a Loom release installed (`loom adamic-env
# install`) carries the marker line below, and no setup run may rewrite it: on Nov 6 a proof's setup run
# overwrote Pupa's, dropped WASI_SYSROOT, unfit her serve and stalled the nightly witness (#zb2bhp9).
# A proof sources the box's env.sh; it never writes it. ADAMIC_SETUP_FORCE=1 overrides, by hand only.
adamic_env_release_marker='# adamic-env: owned by a Loom release'

# Exits 2 when the env.sh given is release-owned and no force is set.
adamic_env_refuse_release_owned() {
	local environment=$1
	if [ "${ADAMIC_SETUP_FORCE:-0}" != 1 ] && [ -f "$environment" ] && grep -qxF "$adamic_env_release_marker" "$environment"; then
		echo "setup: $environment is owned by a Loom release; source it instead (set ADAMIC_SETUP_FORCE=1 to overwrite)" >&2
		exit 2
	fi
}
