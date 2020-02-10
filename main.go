package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	mongowrapper "github.com/opencensus-integrations/gomongowrapper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporter/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const namespace = "github.com/wingyplus/mongo-telemetry"

func main() {
	// opencensus metrics exporter
	initMetrics()
	tracer := initTracer(namespace)

	err := mongowrapper.RegisterAllViews()
	stopIfErr(err, "mongowrapper.RegisterAllViews")

	ctx, span := tracer.Start(context.Background(), "demo")

	client, err := mongowrapper.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	stopIfErr(err, "mongowrapper.NewClient")

	stopIfErr(client.Connect(ctx), "client.Connect")

	col := client.Database("otel").Collection("notes")
	result, err := col.InsertOne(ctx, bson.M{"text": "OTEL works!!"})
	stopIfErr(err, "col.InsertOne")

	fmt.Println(result.InsertedID)

	client.Disconnect(ctx)

	// NOTE(wingyplus): I cannot use defer. Span will end after exporter sync to tracing system and
	// Jaeger will be mark trace as <trace-without-root-span>.
	span.End()

	// NOTE(wingyplus): otel exporter needs more some time to sync with tracing systems.
	time.Sleep(4 * time.Second)
}

func initMetrics() {
	exporter, err := prometheus.NewExporter(prometheus.Options{
		Namespace: namespace,
	})
	stopIfErr(err, "prometheus.NewExporter")
	view.RegisterExporter(exporter)
}

func initTracer(name string) trace.Tracer {
	// exporter, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
	// stopIfErr(err, "stdout.NewExporter")
	exporter, err := jaeger.NewExporter(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(
			jaeger.Process{
				ServiceName: name,
			},
		),
	)
	stopIfErr(err, "jaeger.NewExporter")
	provider, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter),
	)
	stopIfErr(err, "trace.NewProvider")
	global.SetTraceProvider(provider)
	return global.TraceProvider().Tracer(name)
}

func stopIfErr(err error, step string) {
	if err != nil {
		log.Fatalf("%s: have some problems: %v", step, err)
	}
}
