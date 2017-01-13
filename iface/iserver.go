package iface

import (
	"time"
)

type Iserver interface {
	Start()
	Stop()
	AddRouter(router interface{})
	CallLater(durations time.Duration, f func(v ...interface{}), args ...interface{})
	CallWhen(ts string, f func(v ...interface{}), args ...interface{})
	CallLoop(durations time.Duration, f func(v ...interface{}), args ...interface{})
}
