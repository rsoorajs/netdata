// SPDX-License-Identifier: GPL-3.0-or-later

package icecast

import (
	"fmt"

	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/web"
)

const (
	urlPathServerStats = "/status-json.xsl" // https://icecast.org/docs/icecast-trunk/server_stats/
)

func (ic *Icecast) collect() (map[string]int64, error) {
	mx := make(map[string]int64)

	if err := ic.collectServerStats(mx); err != nil {
		return nil, err
	}

	return mx, nil
}

func (ic *Icecast) collectServerStats(mx map[string]int64) error {
	stats, err := ic.queryServerStats()
	if err != nil {
		return err
	}
	if stats.IceStats == nil {
		return fmt.Errorf("unexpected response: no icestats found")
	}
	if len(stats.IceStats.Source) == 0 {
		return fmt.Errorf("no icecast sources found")
	}

	seen := make(map[string]bool)

	for _, src := range stats.IceStats.Source {
		name := src.ServerName
		if name == "" {
			continue
		}

		seen[name] = true

		if !ic.seenSources[name] {
			ic.seenSources[name] = true
			ic.addSourceCharts(name)
		}

		px := fmt.Sprintf("source_%s_", name)

		mx[px+"listeners"] = src.Listeners
	}

	for name := range ic.seenSources {
		if !seen[name] {
			delete(ic.seenSources, name)
			ic.removeSourceCharts(name)
		}
	}

	return nil
}

func (ic *Icecast) queryServerStats() (*serverStats, error) {
	req, err := web.NewHTTPRequestWithPath(ic.RequestConfig, urlPathServerStats)
	if err != nil {
		return nil, err
	}

	var stats serverStats
	if err := web.DoHTTP(ic.httpClient).RequestJSON(req, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}
