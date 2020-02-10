module github.com/wingyplus/mongo-telemetry

go 1.14

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/opencensus-integrations/gomongowrapper v0.0.1
	go.mongodb.org/mongo-driver v1.3.0
	go.opencensus.io v0.22.3
	go.opentelemetry.io/otel v0.2.1
	go.opentelemetry.io/otel/exporter/trace/jaeger v0.2.1
)

replace github.com/opencensus-integrations/gomongowrapper => ../gomongowrapper
