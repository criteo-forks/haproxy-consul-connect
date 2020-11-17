package state

import (
	"github.com/haproxytech/models/v2"
)

func generateStats(opts Options, state State) (State, error) {
	// Main config
	fe := Frontend{
		Frontend: models.Frontend{
			Name:       "stats",
			HTTPUseHtx: models.FrontendHTTPUseHtxEnabled,
			Httplog:    opts.LogRequests,
			StatsOptions: &models.StatsOptions{
				StatsEnable:       true,
				StatsURIPrefix:    "/stats",
				StatsRefreshDelay: int64p(10),
			},
		},
		Bind: models.Bind{
			Name:    "stats_bind",
			Address: "*",
			Port:    int64p(8484),
		},
		HTTPRequestRules: []*models.HTTPRequestRule{
			{},
		},
	}
}
