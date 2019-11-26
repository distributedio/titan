package conf

import "time"

// MockConf init and return titan mock conf
func MockConf() *Titan {
	return &Titan{
		TiKV: TiKV{
			PdAddrs: "mocktikv://",
			GC: GC{
				Disable:        false,
				Interval:       time.Second,
				LeaderLifeTime: 3 * time.Minute,
				BatchLimit:     256,
			},
			Expire: Expire{
				Disable:        false,
				Interval:       time.Second,
				LeaderLifeTime: 3 * time.Minute,
				BatchLimit:     256,
			},
			ZT: ZT{
				Disable:    false,
				Workers:    5,
				BatchCount: 10,
				QueueDepth: 100,
				Interval:   1000 * time.Millisecond,
			},
			TiKVGC: TiKVGC{
				Disable:           false,
				Interval:          20 * time.Minute,
				LeaderLifeTime:    30 * time.Minute,
				SafePointLifeTime: 10 * time.Minute,
				Concurrency:       2,
			},
		},
	}
}
