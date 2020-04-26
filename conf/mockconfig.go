package conf

import "time"

// MockConf init and return titan mock conf
func MockConf() *Titan {
	return &Titan{
		Tikv: Tikv{
			PdAddrs: "mocktikv://",
			GC: GC{
				Disable:        false,
				Interval:       time.Second,
				LeaderLifeTime: 3 * time.Minute,
				BatchLimit:     256,
			},
			Expire: Expire{
				Disable:          false,
				Interval:         time.Second,
				LeaderLifeTime:   3 * time.Minute,
				BatchLimit:       10,
				UnhashBatchLimit: 256,
			},
			ZT: ZT{
				Disable:    false,
				Workers:    5,
				BatchCount: 10,
				QueueDepth: 100,
				Interval:   1000 * time.Millisecond,
			},
			TikvGC: TikvGC{
				Disable:           false,
				Interval:          20 * time.Minute,
				LeaderLifeTime:    30 * time.Minute,
				SafePointLifeTime: 10 * time.Minute,
				Concurrency:       2,
			},
			RateLimit: RateLimit{
				LimiterNamespace:    "sys_ratelimit",
				SyncSetPeriod:       1 * time.Second,
				GlobalBalancePeriod: 3 * time.Second,
				TitanStatusLifetime: 6 * time.Second,
				UsageToDivide:       0.6,
				UsageToMultiply:     0.9,
				WeightChangeFactor:  1.5,
				InitialPercent:      1,
			},
		},
	}
}
