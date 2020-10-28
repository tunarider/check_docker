package nagios

type State int

func (s State) Int() int {
	return int(s)
}

func (s State) Gt(c State) State {
	if s.Int() > c.Int() {
		return s
	} else {
		return c
	}
}

const (
	StateOk       State = 0
	StateWarning  State = 1
	StateCritical State = 2
	StateUnknown  State = 3
)

func ResolveState(states ...State) State {
	o := StateOk
	for _, state := range states {
		o = o.Gt(state)
	}
	return o
}
