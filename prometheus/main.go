package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

const meterName = "example/prometheus"

func main() {
	//用当前时间戳初始化一个随机数
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctx := context.Background()
	//	导出器嵌入默认的 OpenTelemetry Reader，并且 实现 Prometheus。收集器，允许将其用作
	//	既是读者又是收集者。
	exporter, err := prometheus.New() // 生成http客户端
	if err != nil {
		log.Fatal(err)
	}
	// 创建了一个新的度量提供者（Meter Provider），并配置了一个读取器。
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	// Meter是创建和记录度量指标的主要接口
	meter := provider.Meter(meterName)
	//启动 prometheus HTTP 服务器并将导出器 Collector 传递给它
	// Start the prometheus HTTP server and pass the exporter Collector to it
	go serveMetrics()
	// 定义了一组属性（Attributes）用于附加到度量指标上，增加度量数据的上下文信息。
	opt := api.WithAttributes(
		attribute.Key("A").String("B"),
		attribute.Key("C").String("D"),
	)
	//  创建计数器,第一个参数设置名称，用于搜索，第二个参数，
	// This is the equivalent of prometheus.NewCounterVec
	counter, err := meter.Float64Counter("foo", api.WithDescription("a simple counter"))
	if err != nil {
		log.Fatal(err)
	}
	counter.Add(ctx, 5, opt)
	// 创建可观测性量表
	gauge, err := meter.Float64ObservableGauge("bar", api.WithDescription("a fun little gauge"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = meter.RegisterCallback(func(_ context.Context, o api.Observer) error {
		n := -10. + rng.Float64()*(90.) // [-10, 100)
		o.ObserveFloat64(gauge, n, opt)
		return nil
	}, gauge)
	if err != nil {
		log.Fatal(err)
	}
	// 创建直方图
	// This is the equivalent of prometheus.NewHistogramVec
	histogram, err := meter.Float64Histogram(
		"baz",
		api.WithDescription("a histogram with custom buckets and rename"),
		api.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)
	if err != nil {
		log.Fatal(err)
	}
	histogram.Record(ctx, 136, opt)
	histogram.Record(ctx, 64, opt)
	histogram.Record(ctx, 701, opt)
	histogram.Record(ctx, 830, opt)

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	<-ctx.Done()
}

func serveMetrics() {
	log.Printf("serving metrics at localhost:8080/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8080", nil) //nolint:gosec // Ignoring G114: Use of net/http serve function that has no support for setting timeouts.
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
