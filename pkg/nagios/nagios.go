package nagios

type State int

const (
	StateOk       State = 0
	StateWarning  State = 1
	StateCritical State = 2
	StateUnknown  State = 3
)

func StateToInt(s State) int {
	return int(s)
}

func StateGt(s State, c State) State {
	if StateToInt(s) > StateToInt(c) {
		return s
	} else {
		return c
	}
}

func ResolveState(states ...State) State {
	o := StateOk
	for _, state := range states {
		o = StateGt(o, state)
	}
	return o
}
