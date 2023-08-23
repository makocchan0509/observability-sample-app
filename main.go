package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

//var tracer trace.Tracer

func main() {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	name := os.Getenv("APP_NAME")
	port := os.Getenv("APP_PORT")

	var endpoints []string
	e := os.Getenv("ENDPOINTS")
	if e != "" {
		endpoints = strings.Split(e, ",")
	}

	// ハンドラー関数を定義する
	proxy := func(w http.ResponseWriter, _ *http.Request) {

		apiRes := []string{name}

		for i := 0; i < len(endpoints); i++ {

			cli, err := newHttpClient(endpoints[i], "GET")
			if err != nil {
				log.Error().Msgf("http client error: %v\n", err)
				w.WriteHeader(500)
				continue
			}
			cli.SendRequest()
			res, err := cli.RespToString()
			if err != nil {
				log.Error().Msgf("http client error: %v\n", err)
				res = "err"
			}
			sp := strings.Split(res, "-")
			for j := 0; j < len(sp); j++ {
				apiRes = append(apiRes, sp[j])
			}
			cli.Close()
		}
		io.WriteString(w, fmt.Sprintf("%s\n", strings.Join(apiRes, "-")))
	}

	// パスとハンドラー関数を結びつける
	http.HandleFunc("/", proxy)
	http.Handle("/metrics", promhttp.Handler())

	// Web サーバーを起動する
	log.Info().Msg("Starting server...")
	log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

//func observMiddleware(next http.Handler) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//
//		next.ServeHTTP(w, r)
//	}
//}

type HttpClient struct {
	req    *http.Request
	client *http.Client
	method string
	url    string
	resp   *http.Response
}

func newHttpClient(url string, method string) (*HttpClient, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return &HttpClient{}, err
	}
	client := new(http.Client)
	return &HttpClient{
		req:    req,
		client: client,
		method: method,
		url:    url,
	}, nil
}

func (h *HttpClient) setHeader(header map[string]string) {
	for k, v := range header {
		h.req.Header.Set(k, v)
	}
}

func (h *HttpClient) SendRequest() error {
	r, err := h.client.Do(h.req)
	if err != nil {
		return err
	}
	h.resp = r
	return nil
}

func (h *HttpClient) Close() {
	h.resp.Body.Close()
}

func (h *HttpClient) RespToString() (string, error) {
	b, err := ioutil.ReadAll(h.resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//func initTracer() func() {
//	ctx := context.Background()
//
//	driver := otlpgrpc.NewDriver(
//		otlpgrpc.WithInsecure(),
//		otlpgrpc.WithEndpoint("tempo:55680"),
//		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
//	)
//	exp, err := otlp.NewExporter(ctx, driver)
//	handleErr(err, "failed to create exporter")
//
//	res, err := resource.New(ctx,
//		resource.WithAttributes(
//			// the service name used to display traces in backends
//			semconv.ServiceNameKey.String("demo-service"),
//		),
//	)
//	handleErr(err, "failed to create resource")
//
//	bsp := sdktrace.NewBatchSpanProcessor(exp)
//	tracerProvider := sdktrace.NewTracerProvider(
//		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
//		sdktrace.WithResource(res),
//		sdktrace.WithSpanProcessor(bsp),
//	)
//
//	// set global propagator to tracecontext (the default is no-op).
//	otel.SetTextMapPropagator(propagation.TraceContext{})
//	otel.SetTracerProvider(tracerProvider)
//
//	return func() {
//		// Shutdown will flush any remaining spans.
//		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")
//	}
//}
