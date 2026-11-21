set -u
# #jjxtwzs: the four programs the measurement saw move, each compiled at head with its whole stop and its set.
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
go build -o "$TMPDIR/adamic" ./cmd/adamic || { echo "moved: FAIL build"; exit 1; }
for program in internal/oracle/testdata/node_process_handle_detached_key.a internal/oracle/testdata/node_process_handle_optional.a internal/oracle/testdata/records_pick_views.a internal/oracle/testdata/view_array_object_stops/field_empty_interface.ts; do
	echo "=================== $program"
	set_file="$TMPDIR/set.tsv"; : > "$set_file"
	ADAMIC_SCRATCH_ARRAYS_SET="$set_file" "$TMPDIR/adamic" c "$program" > /dev/null 2> "$TMPDIR/err.log"; echo "moved: exit=$?"
	head -c 1200 "$TMPDIR/err.log"; echo
	echo "moved: set: $(tail -1 "$set_file" | cut -f2- | cut -c1-400)"
	grep -n "setBlocking\|fd?\|Pick<\|as \|: Options\|_handle" "$program" | head -12
done
echo MOVED DONE
