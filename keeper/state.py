"""state.py: every edit to the fanout's state.json, under one lock, so the dispatcher, the ramp and the keeper never
write over each other (a lost write dropped two running units on Oct 9, 03:12).

	state.py <state.json> take                    print the next queued unit and remove it ("" when empty)
	state.py <state.json> run <clone> <unit>      record the clone as running the unit
	state.py <state.json> unit <clone>            print the unit the clone runs ("" when none)
	state.py <state.json> finish <clone>          the clone's unit is done
	state.py <state.json> requeue <clone>         the clone's unit goes back to the front of the queue
	state.py <state.json> put-back <unit>         a unit taken but never started goes back to the front
	state.py <state.json> left                    queued plus running
	state.py <state.json> drop <unit>...          take units out of the queue for good (not units: a testdata side-file directory)
"""
import fcntl, json, os, sys

path, verb, *arguments = sys.argv[1:]
with open(path + ".lock", "w") as lock:
    fcntl.flock(lock, fcntl.LOCK_EX)
    state = json.load(open(path))
    answer, changed = "", True
    if verb == "take":
        answer = state["queue"].pop(0) if state["queue"] else ""
    elif verb == "run":
        state["running"][arguments[0]] = arguments[1]
    elif verb == "unit":
        answer, changed = state["running"].get(arguments[0], ""), False
    elif verb == "finish":
        unit = state["running"].pop(arguments[0], "")
        if unit:
            state["done"].append(unit)
        answer = unit
    elif verb == "requeue":
        unit = state["running"].pop(arguments[0], "")
        if unit:
            state["queue"].insert(0, unit)
        answer = unit
    elif verb == "put-back":
        state["queue"].insert(0, arguments[0])
    elif verb == "drop":
        state["queue"] = [unit for unit in state["queue"] if unit not in arguments]
        state.setdefault("dropped", []).extend(arguments)
        answer = str(len(state["queue"]))
    elif verb == "left":
        answer, changed = str(len(state["queue"]) + len(state["running"])), False
    else:
        sys.exit("state.py: no verb " + verb)
    if changed:
        json.dump(state, open(path + ".partial", "w"), indent=1)
        os.rename(path + ".partial", path)
print(answer)
