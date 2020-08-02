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
	"math"
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

const (
	name   = "prometheus_ascii"
	layout = "2006-01-02T15:04:05Z"
)

func queryPrometheus(promQuery string, server string, start time.Time, end time.Time, step time.Duration, width int, height int, logger log.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := prometheus.NewClient(prometheus.Config{Address: server})
	if err != nil {
		level.Error(logger).Log("msg", "Client error", "err", err)
		os.Exit(1)
	}

	q := v1.NewAPI(client)

	value, warn, err := q.QueryRange(ctx, promQuery, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
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

	level.Info(logger).Log("msg", "Retrieved query result", "type", queryType.String())

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
			graph := asciigraph.Plot(data, asciigraph.Caption(promQuery), asciigraph.Width(width), asciigraph.Height(height))
			fmt.Println(graph)
		}
	default:
		level.Error(logger).Log("msg", "Query error", "err", "ValueType of Query Result is unknown")
		os.Exit(1)
	}
}

func main() {
	var (
		promServer = kingpin.Flag("prometheus.server", "Prometheus server.").Envar("PROMETHEUS_SERVER").Required().String()
		promQuery  = kingpin.Flag("prometheus.query", "Prometheus query to submit.").Envar("PROMETHEUS_QUERY").Required().String()
		qDur       = kingpin.Flag("prometheus.query.duration", "Duration of query. Overwritten by start.").Envar("PROMETHEUS_QUERY_DUR").Default("24h").Duration()
		qStart     = kingpin.Flag("prometheus.query.start", "Start time for query. Layout: 2006-01-02T15:04:05Z").Envar("PROMETHEUS_QUERY_START").String()
		qEnd       = kingpin.Flag("prometheus.query.end", "End time for query. Defaults to now. Layout: 2006-01-02T15:04:05Z").Envar("PROMETHEUS_QUERY_END").String()
		gWidth     = kingpin.Flag("prometheus.graph.width", "Width of the graph.").Envar("GRAPH_WIDTH").Default("100").Int()
		gHeight    = kingpin.Flag("prometheus.graph.height", "Height of the graph.").Envar("GRAPH_HEIGHT").Default("10").Int()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print(name))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting", "version", version.Info())

	var qEndTime time.Time
	var qEndTimeErr error
	if *qEnd != "" {
		qEndTime, qEndTimeErr = time.Parse(layout, *qEnd)
		if qEndTimeErr != nil {
			level.Error(logger).Log("msg", "Could not parse end time", "err", qEndTimeErr)
			os.Exit(1)
		}
	} else {
		qEndTime = time.Now()
	}

	var qStartTime time.Time
	var qStartTimeErr error
	if *qStart != "" {
		qStartTime, qStartTimeErr = time.Parse(layout, *qStart)
		if qStartTimeErr != nil {
			level.Error(logger).Log("msg", "Could not parse start time", "err", qStartTimeErr)
			os.Exit(1)
		}
	} else {
		qStartTime = qEndTime.Add(-*qDur)
	}

	queryDuration := qEndTime.Sub(qStartTime)
	level.Info(logger).Log("msg", "Query duration", "seconds", queryDuration.Seconds())

	calcQueryStep := math.Round(queryDuration.Seconds() / float64(*gWidth))
	calcQueryStepDur := time.Second * time.Duration(calcQueryStep)
	level.Info(logger).Log("msg", "Calculated step", "seconds", calcQueryStepDur.Seconds())

	queryPrometheus(*promQuery, *promServer, qStartTime, qEndTime, calcQueryStepDur, *gWidth, *gHeight, logger)
}
