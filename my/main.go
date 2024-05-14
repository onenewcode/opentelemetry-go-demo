package app

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"log"
)

var tracer trace.Tracer

func newExporter(ctx context.Context) (*jaeger.Exporter, error) /* (someExporter.Exporter, error) */ {
	// 你选择的 exporter：console、jaeger、zipkin、OTLP 等等。

	// 让jaeger作为我们的后端

	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// 确保默认的 SDK 资源和所需的服务名称已设置。
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("ExampleService"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

func main() {
	ctx := context.Background()

	exp, err := newExporter(ctx)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	// 创建一个跟踪器提供程序，并使用给定的 exporter、批量 span 处理器。
	tp := newTraceProvider(exp)

	// 正确处理关闭操作以避免资源泄漏。
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)

	// 最后，设置可用于该包的跟踪器。
	tracer = tp.Tracer("ExampleService")
}
