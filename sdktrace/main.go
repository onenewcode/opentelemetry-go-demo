package main

import (
	"context"
	"fmt"
	"github.com/go-logr/stdr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"log"
)

var (
	fooKey     = attribute.Key("ex.com/foo")
	barKey     = attribute.Key("ex.com/bar")
	anotherKey = attribute.Key("ex.com/another")
)

var lemonsKey = attribute.Key("ex.com/lemons")

// SubOperation是演示命名跟踪程序使用的示例。
// 它创建一个带有包路径的命名跟踪程序。
func SubOperation(ctx context.Context) error {
	// 创建一个追踪器
	tr := otel.Tracer("go.opentelemetry.io/otel/example/namedtracer/foo")
	// 创建一个span
	var span trace.Span
	_, span = tr.Start(ctx, "Sub operation...")
	defer span.End()
	span.SetAttributes(lemonsKey.String("five"))
	span.AddEvent("Sub span event")
	return nil
}

// 追踪器全局变量，以供其他模块使用
var tp *sdktrace.TracerProvider

// initTracer创建并注册跟踪提供程序实例。
func initTracer() error {

	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return fmt.Errorf("failed to initialize stdouttrace exporter: %w", err)
	}
	// 这里设置标准exporter，除此之外还有ja和otel
	//NewBatchSpanProcessor创建一个新的SpanProcessor，它将使用提供的选项将完成的span批发送给导出器。
	bsp := sdktrace.NewBatchSpanProcessor(exp)

	tp = sdktrace.NewTracerProvider(
		// 设置采样器，这里设置成AlwaysSample，表示所有span都会被采样
		sdktrace.WithSampler(sdktrace.AlwaysSample()),

		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
	return nil
}

func main() {
	//将日志级别设置为info以查看SDK状态消息
	stdr.SetVerbosity(5)

	if err := initTracer(); err != nil {
		log.Panic(err)
	}

	// 创建一个以包路径作为其名称的命名跟踪程序。
	tracer := tp.Tracer("go.opentelemetry.io/otel/example/namedtracer")
	ctx := context.Background()
	defer func() { _ = tp.Shutdown(ctx) }()
	m0, err := baggage.NewMemberRaw(string(fooKey), "foo1")
	if err != nil {
		log.Println("failed to create Member m0")
		return
	}
	m1, err := baggage.NewMemberRaw(string(barKey), "bar1")
	if err != nil {
		log.Println("failed to create Member m1")
		return
	}
	b, err := baggage.New(m0, m1)
	if err != nil {
		log.Println("failed to create baggage")
		return
	}
	ctx = baggage.ContextWithBaggage(ctx, b)

	var span trace.Span
	ctx, span = tracer.Start(ctx, "operation")
	defer span.End()
	span.AddEvent("Nice operation!", trace.WithAttributes(attribute.Int("bogons", 100)))
	span.SetAttributes(anotherKey.String("yes"))
	if err := SubOperation(ctx); err != nil {
		panic(err)
	}
}
