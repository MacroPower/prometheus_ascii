/*
Prometheus ASCII Client
Copyright (C) 2020 Jacob Colvin (MacroPower)

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	prometheus "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/guptarohit/asciigraph"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

func queryPrometheus(promQuery string, server string, offset time.Duration, logger log.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := prometheus.NewClient(prometheus.Config{Address: server})
	if err != nil {
		level.Error(logger).Log("msg", "Client error", "err", err)
		os.Exit(1)
	}

	q := v1.NewAPI(client)

	value, warn, err := q.QueryRange(ctx, promQuery, v1.Range{
		Start: time.Now().Add(offset),
		End:   time.Now(),
		Step:  time.Minute,
	})
	if len(warn) > 0 {
		level.Error(logger).Log("msg", "Too many warnings for query", "query", promQuery, "warn", warn)
		os.Exit(1)
	}
	if err != nil {
		level.Error(logger).Log("msg", "Query error", "err", err)
		os.Exit(1)
	}

	queryType := value.Type()

	fmt.Println(queryType.String())

	switch {
	case queryType == model.ValScalar:
		scalarVal := value.(*model.Scalar)
		fmt.Println(scalarVal.Value)
	case queryType == model.ValVector:
		vectorVal := value.(model.Vector)
		for _, elem := range vectorVal {
			fmt.Println(elem.Value)
		}
	case queryType == model.ValMatrix:
		matrixVal := value.(model.Matrix)
		for _, series := range matrixVal {
			data := []float64{}
			for _, elem := range series.Values {
				data = append(data, float64(elem.Value))
			}
			graph := asciigraph.Plot(data, asciigraph.Caption(promQuery), asciigraph.Width(100), asciigraph.Height(10))
			fmt.Println(graph)
		}
	default:
		level.Error(logger).Log("msg", "Query error", "err", "ValueType of Query Result is unknown")
		os.Exit(1)
	}
}

func main() {
	var (
		promQuery = kingpin.Flag("prometheus.query", "Prometheus query to submit.").Required().String()
		server    = kingpin.Flag("prometheus.server", "Prometheus server.").Required().String()
		start     = kingpin.Flag("query.start", "Time offset (from now) for query.").Default("-30m").Duration()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("prometheus_ascii"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting", "version", version.Info())

	queryPrometheus(*promQuery, *server, *start, logger)
}
