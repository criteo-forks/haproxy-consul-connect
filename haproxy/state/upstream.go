package state

import (
	"fmt"

	"github.com/haproxytech/haproxy-consul-connect/consul"
	"github.com/haproxytech/models/v2"
)

func generateUpstream(opts Options, certStore CertificateStore, cfg consul.Upstream, oldState, newState State) (State, error) {
	feName := fmt.Sprintf("front_%s", cfg.Name)
	beName := fmt.Sprintf("back_%s", cfg.Name)
	feMode := models.FrontendModeHTTP
	beMode := models.BackendModeHTTP

	fePort64 := int64(cfg.LocalBindPort)

	if cfg.Protocol != "" && cfg.Protocol == "tcp" {
		feMode = models.FrontendModeTCP
		beMode = models.BackendModeTCP
	}

	fe := Frontend{
		Frontend: models.Frontend{
			Name:           feName,
			DefaultBackend: beName,
			ClientTimeout:  int64p(int(cfg.ReadTimeout.Milliseconds())),
			Mode:           feMode,
			Httplog:        opts.LogRequests,
		},
		Bind: models.Bind{
			Name:    fmt.Sprintf("%s_bind", feName),
			Address: cfg.LocalBindAddress,
			Port:    &fePort64,
		},
		FilterCompression: &FrontendFilter{
			Filter: models.Filter{
				Index:      int64p(0),
				Type:       models.FilterTypeCompression,
			},
		},
	}
	if opts.LogRequests && opts.LogSocket != "" {
		fe.LogTarget = &models.LogTarget{
			Index:    int64p(0),
			Address:  opts.LogSocket,
			Facility: models.LogTargetFacilityLocal0,
			Format:   models.LogTargetFormatRfc5424,
		}
	}

	// Connect name header
	if opts.ResponseHdrName != "" && feMode == models.FrontendModeHTTP {
		fmt.Printf("header setting upstream")
		fe.HTTPResponseRules = append(fe.HTTPResponseRules, models.HTTPResponseRule{
			Index:     int64p(0),
			Type:      models.HTTPResponseRuleTypeSetHeader,
			HdrName:   opts.ResponseHdrName,
			HdrFormat: "true",
		})
	}

	newState.Frontends = append(newState.Frontends, fe)

	be := Backend{
		Backend: models.Backend{
			Name:           beName,
			ServerTimeout:  int64p(int(cfg.ReadTimeout.Milliseconds())),
			ConnectTimeout: int64p(int(cfg.ConnectTimeout.Milliseconds())),
			Balance: &models.Balance{
				Algorithm: stringp(models.BalanceAlgorithmLeastconn),
			},
			Mode: beMode,
		},
	}
	if opts.LogRequests && opts.LogSocket != "" {
		be.LogTarget = &models.LogTarget{
			Index:    int64p(0),
			Address:  opts.LogSocket,
			Facility: models.LogTargetFacilityLocal0,
			Format:   models.LogTargetFormatRfc5424,
		}
	}

	servers, err := generateUpstreamServers(opts, certStore, cfg, beName, oldState)
	if err != nil {
		return newState, err
	}
	be.Servers = servers
	newState.Backends = append(newState.Backends, be)

	return newState, nil
}

func generateUpstreamServers(opts Options, certStore CertificateStore, cfg consul.Upstream, beName string, oldState State) ([]models.Server, error) {
	oldBackend, _ := oldState.findBackend(beName)

	idxHANode := func(s models.Server) string {
		if s.Maintenance == models.ServerMaintenanceEnabled {
			return "maint"
		}
		return fmt.Sprintf("%s:%d", s.Address, *s.Port)
	}
	idxConsulNode := func(s consul.UpstreamNode) string {
		return fmt.Sprintf("%s:%d", s.Host, s.Port)
	}

	servers := make([]models.Server, len(oldBackend.Servers))
	copy(servers, oldBackend.Servers)
	serversIdx := index(servers, func(i int) string {
		return idxHANode(servers[i])
	})

	newServersIdx := index(cfg.Nodes, func(i int) string {
		return idxConsulNode(cfg.Nodes[i])
	})

	caPath, crtPath, err := certStore.CertsPath(cfg.TLS)
	if err != nil {
		return nil, err
	}

	disabledServer := models.Server{
		Address:        "127.0.0.1",
		Port:           int64p(1),
		Weight:         int64p(1),
		Ssl:            models.ServerSslEnabled,
		SslCertificate: crtPath,
		SslCafile:      caPath,
		Verify:         models.BindVerifyRequired,
		Maintenance:    models.ServerMaintenanceEnabled,
	}

	emptyServerSlots := make([]int, 0, len(servers))

	// Disable removed servers
	for i, s := range servers {
		_, ok := newServersIdx[idxHANode(s)]
		if ok {
			continue
		}

		servers[i] = disabledServer
		servers[i].Name = fmt.Sprintf("srv_%d", i)
		emptyServerSlots = append(emptyServerSlots, i)
	}

	// Add new servers
	for _, s := range cfg.Nodes {
		i, ok := serversIdx[idxConsulNode(s)]
		if ok {
			// if the server exists, just update its certificate in case they changed
			servers[i].SslCafile = caPath
			servers[i].SslCertificate = crtPath
			continue
		}

		if len(emptyServerSlots) == 0 {
			l := len(servers)
			add := l
			if add == 0 {
				add = 1
			}
			for i := 0; i < add; i++ {
				server := disabledServer
				server.Name = fmt.Sprintf("srv_%d", i+l)
				servers = append(servers, server)
				emptyServerSlots = append(emptyServerSlots, i+l)
			}
		}

		i = emptyServerSlots[0]
		emptyServerSlots = emptyServerSlots[1:]

		servers[i].Address = s.Host
		servers[i].Port = int64p(s.Port)
		servers[i].Weight = int64p(s.Weight)
		servers[i].Maintenance = models.ServerMaintenanceDisabled
	}

	return servers, nil
}
