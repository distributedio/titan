package conf

import "time"

// MockConf init and return titan mock conf
func MockConf() *Titan {
	return &Titan{
		Tikv: Tikv{
			PdAddrs: "mocktikv://",
			GC: GC{
				Enable:         true,
				Interval:       time.Second,
				LeaderLifeTime: 3 * time.Minute,
				BatchLimit:     256,
			},
			Expire: Expire{
				Enable:         true,
				Interval:       time.Second,
				LeaderLifeTime: 3 * time.Minute,
				BatchLimit:     256,
			},
			ZT: ZT{
				Enable:     true,
				Workers:    5,
				BatchCount: 10,
				QueueDepth: 100,
				Interval:   1000 * time.Millisecond,
			},
			TikvGC: TikvGC{
				Enable:            true,
				Interval:          20 * time.Minute,
				LeaderLifeTime:    30 * time.Minute,
				SafePointLifeTime: 10 * time.Minute,
				Concurrency:       2,
			},
		},
	}
}
