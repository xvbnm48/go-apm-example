package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func SetUpTracer() (*tracesdk.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-apm-example"),
		)),
	)
	return tp, nil

}
func main() {
	tp, err := SetUpTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}

	defer func() {
		if er := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", er)
		}
	}()

	otel.SetTracerProvider(tp)
	// handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprint(w, "Hello, World!")
	// })
	handler := http.HandlerFunc(HelloHandler)

	wraphandler := otelhttp.NewHandler(handler, "hello-handler")
	http.Handle("/", wraphandler)
	fmt.Println("Server started at :8080")
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, "Hello, World! %d\n", i)
	}
}
