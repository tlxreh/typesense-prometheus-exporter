package typesense_prometheus_exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type TypesenseCollector struct {
	ctx        context.Context
	logger     *slog.Logger
	endPoint   string
	apiKey     string
	cluster    string
	httpClient *http.Client
	metrics    map[string]*prometheus.Desc
	stats      map[string]*prometheus.Desc
	mutex      sync.Mutex
}

var (
	metricLabels   = []string{"typesense_cluster"}
	endpointLabels = []string{"typesense_cluster", "typesense_request"}
)

func NewTypesenseCollector(ctx context.Context, logger *slog.Logger, config Config) *TypesenseCollector {
	collector := &TypesenseCollector{
		ctx:      ctx,
		logger:   logger,
		endPoint: fmt.Sprintf("%s://%s:%d", config.Protocol, config.Host, config.ApiPort),
		apiKey:   config.ApiKey,
		cluster:  config.Cluster,
		httpClient: &http.Client{
			Timeout: 500 * time.Millisecond,
		},
		metrics: getMetricsDesc(),
		stats:   getStatsDesc(),
	}

	return collector
}

// Describe sends the metric descriptors to the Prometheus channel
func (c *TypesenseCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}

	for _, stat := range c.stats {
		ch <- stat
	}
}

// Collect fetches the metrics from the Typesense endpoint and sends them to the Prometheus channel
func (c *TypesenseCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	c.logger.Info(fmt.Sprintf("collecting data"), "cluster", c.cluster, "endpoint", c.endPoint)

	c.mutex.Lock()
	defer func() {
		elapsed := time.Since(start)
		c.logger.Info(fmt.Sprintf("collecting data completed"), "duration", elapsed)
		c.mutex.Unlock()
	}()

	targets := []string{"metrics", "stats"}
	for _, target := range targets {
		data, err := c.fetch(target)
		if err != nil {
			return
		}

		c.collect(target, data, ch)
	}
}

func (c *TypesenseCollector) fetch(target string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s.json", c.endPoint, target)
	c.logger.Info(fmt.Sprintf("collecting %s", target), "cluster", c.cluster, "url", url)

	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error(fmt.Sprintf("error creating request: %v", err))
		return nil, err
	}

	req.Header.Set("x-typesense-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error(fmt.Sprintf("error fetching %s: %v", target, err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error(fmt.Sprintf("error fetching %s: %v", target, resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(fmt.Sprintf("error reading response body from %s: %v", url, err))
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.logger.Error(fmt.Sprintf("error unmarshalling %s.json body: %v", target, err))
		return nil, err
	}

	return data, nil
}

func (c *TypesenseCollector) collect(target string, data map[string]interface{}, ch chan<- prometheus.Metric) {
	for key, value := range data {
		select {
		case <-c.ctx.Done():
			c.logger.Error(fmt.Sprintf("context canceled, stopping collection"))
			return
		default:
		}

		switch target {
		case "metrics":
			if desc, ok := c.metrics[key]; ok {
				if sval, ok := value.(string); ok {
					val, err := strconv.ParseFloat(sval, 64)
					if err != nil {
						c.logger.Error(fmt.Sprintf("error converting value for %s: %v", key, err))
						continue
					}
					metric := prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, c.cluster)
					c.logger.Debug(fmt.Sprintf("collected %s", target), "key", key, "value", val)

					ch <- metric
				}
			}
		case "stats":
			if nestedData, ok := data[key]; ok {
				if endpoints, ok := nestedData.(map[string]interface{}); ok {
					for endpoint, endpointVal := range endpoints {
						if desc, ok := c.stats[key]; ok {
							if val, ok := endpointVal.(float64); ok {
								stat := prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, c.cluster, endpoint)
								c.logger.Debug(fmt.Sprintf("collected %s", target), "key", key, "endpoint", endpoint, "value", val)

								ch <- stat
							}
						}
					}
				} else {
					if desc, ok := c.stats[key]; ok {
						if val, ok := value.(float64); ok {
							stat := prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, c.cluster)
							c.logger.Debug(fmt.Sprintf("collected %s", target), "key", key, "value", val)

							ch <- stat
						}
					}
				}
			}
		}
	}
}
