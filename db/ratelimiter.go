package db

import (
	"sync"
	"time"
)

const (
	ALL_NAMESPACE = "*"
)

type CommandLimiter struct {
	lock                       sync.Mutex
	startTime                  time.Time
	commandsCount              int
	commandsSize               int
	maxCommandsCount           int
	maxCommandsSize            int
	period                     int
	periodsNumToTriggerBalance int
	totalCommandsCount         int
	totalCommandsSize          int
}

type NamespaceLimiter struct {
	lock           sync.Mutex
	commandLimiter map[string]*CommandLimiter
}

type LimitRule struct {
	//namespace string
	//command string
	maxqps  int
	maxrate int
}
type Limiter struct {
	lock sync.Mutex

	limitPeriod   int
	balancePeriod int
	currentIp     string

	limitRules map[string]map[string]LimitRule
	limiter    map[string]NamespaceLimiter
}

func NewLimitRule(rule string) map[string]map[string]LimitRule {
	return nil
}

func NewLimiter(limitPeriod int, balancePeriod int, rule string) *Limiter {
	l := &Limiter{
		limitPeriod:   limitPeriod,
		balancePeriod: balancePeriod,
		currentIp:     "127.0.0.1", //set current ip
		limitRules:    NewLimitRule(rule),
		limiter:       make(map[string]NamespaceLimiter),
	}

	return l
}

func (l *Limiter) addAllNamespaceLimiter(rule string) {

}

//func (l *Limiter) addNamespaceLimiter(namespace string, command string, maxQps int, maxRate int) {
//
//}

func (l *Limiter) get(namespace string, command string) *CommandLimiter {
	return nil
}

func (l *Limiter) setCommandLimiter(namespace string, command string, commandLimiter *CommandLimiter) *CommandLimiter {
	return nil
}

func (l *Limiter) checkLimit(namespace string, command string, argSize int) {
	commandLimiter := l.get(namespace, command)
	if commandLimiter == nil {
		//get rule from limitRule
		limitRule, ok := l.limitRules[namespace][command]
		if ok {
			tmp := NewCommandLimiter(limitRule)
			commandLimiter = l.setCommandLimiter(namespace, command, tmp)
		} else {
			limitRule, ok = l.limitRules[ALL_NAMESPACE][command]
			if ok {
				tmp := NewCommandLimiter(limitRule)
				commandLimiter = l.setCommandLimiter(namespace, command, tmp)
			}
		}
	}
	if commandLimiter != nil {
		commandLimiter.checkLimit(len(command), argSize)
	}
}

func (l *Limiter) watchConfigChanged() {

}

func (l *Limiter) clearLimitPeriod() {

}

func (l *Limiter) balanceLimitInNodes() {

}

//func NewNamespaceLimiter() *NamespaceLimiter {
//
//}

func (nl *NamespaceLimiter) getOrSetCommandLimiter(command string, maxQps int, maxRate int) *CommandLimiter {
	return nil
}

//func (nl *NamespaceLimiter) checkLimit (command string, argSize int) {
//
//}

func NewCommandLimiter(limitRule LimitRule) *CommandLimiter {
	return nil
}

func (cl *CommandLimiter) checkLimit(commandSize int, argSize int) {

}

func (cl *CommandLimiter) reset() {

}
