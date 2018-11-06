package command

// Constraint is the rule of command
type Constraint struct {
	Arity    int // number of arguments, it is possible to use -N to say >= N
	Flags    Flag
	FirstKey int
	LastKey  int
	KeyStep  int
}

// Flag is the redis command flag
type Flag int

// Command flags
const (
	CmdWrite Flag = 1 << iota
	CmdReadOnly
	CmdDenyOOM
	CmdModule
	CmdAdmin
	CmdPubsub
	CmdNoScript
	CmdRandom
	CmdSortForScript
	CmdLoading
	CmdStale
	CmdSkipMonitor
	CmdAsking
	CmdFast
	CmdModuleGetKeys
	CmdModuleNoCluster
)

// String returns the string representation of flag
func (f Flag) String() string {
	switch f {
	case CmdWrite:
		return "write"
	case CmdReadOnly:
		return "readonly"
	case CmdDenyOOM:
		return "denyoom"
	case CmdModule:
		return "module"
	case CmdAdmin:
		return "admin"
	case CmdPubsub:
		return "pubsub"
	case CmdNoScript:
		return "noscript"
	case CmdRandom:
		return "random"
	case CmdSortForScript:
		return "sort_for_script"
	case CmdLoading:
		return "loading"
	case CmdStale:
		return "stale"
	case CmdSkipMonitor:
		return "skip_monitor"
	case CmdAsking:
		return "asking"
	case CmdFast:
		return "fast"
	case CmdModuleGetKeys:
		return "module_getkeys"
	case CmdModuleNoCluster:
		return "module_no_cluster"
	}
	return ""
}

// flags parse sflags to flags
// This is the meaning of the flags:
//
//  w: write command (may modify the key space).
//  r: read command  (will never modify the key space).
//  m: may increase memory usage once called. Don't allow if out of memory.
//  a: admin command, like SAVE or SHUTDOWN.
//  p: Pub/Sub related command.
//  f: force replication of this command, regardless of server.dirty.
//  s: command not allowed in scripts.
//  R: random command. Command is not deterministic, that is, the same command
//     with the same arguments, with the same key space, may have different
//     results. For instance SPOP and RANDOMKEY are two random commands.
//  S: Sort command output array if called from script, so that the output
//     is deterministic.
//  l: Allow command while loading the database.
//  t: Allow command while a slave has stale data but is not allowed to
//     server this data. Normally no command is accepted in this condition
//     but just a few.
//  M: Do not automatically propagate the command on MONITOR.
//  k: Perform an implicit ASKING for this command, so the command will be
//     accepted in cluster mode if the slot is marked as 'importing'.
//  F: Fast command: O(1) or O(log(N)) command that should never delay
//     its execution as long as the kernel scheduler is giving us time.
//     Note that commands that may trigger a DEL as a side effect (like SET)
//     are not fast commands.
func flags(s string) Flag {
	flags := Flag(0)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case 'w':
			flags |= CmdWrite
		case 'r':
			flags |= CmdReadOnly
		case 'm':
			flags |= CmdDenyOOM
		case 'a':
			flags |= CmdAdmin
		case 'p':
			flags |= CmdPubsub
		case 's':
			flags |= CmdNoScript
		case 'R':
			flags |= CmdRandom
		case 'S':
			flags |= CmdSortForScript
		case 'l':
			flags |= CmdLoading
		case 't':
			flags |= CmdStale
		case 'M':
			flags |= CmdSkipMonitor
		case 'k':
			flags |= CmdAsking
		case 'F':
			flags |= CmdFast
		default:
			panic("Unsupported command flag")
		}
	}
	return flags
}
func parseFlags(flags Flag) []string {
	var s []string
	// we have total 16 flags now
	for i := uint(0); i < 16; i++ {
		f := Flag(1 << i)
		if f&flags != 0 {
			s = append(s, f.String())
		}
	}
	return s
}
