package ai

import "sync/atomic"

var enabled atomic.Bool

func SetEnabled(on bool) { enabled.Store(on) }

func Enabled() bool { return enabled.Load() }
