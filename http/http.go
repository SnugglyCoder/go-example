package http

import (
	"encoding/json"
	"io"
	http "net/http"

	"github.com/SnugglyCoder/go-example/errs"
	"github.com/SnugglyCoder/go-example/foo"
)

func NewServer(attachments ServerAttachments) http.Server {
	router := http.NewServeMux()
	router.HandleFunc("POST /foo/v1", newHandleFooV1Post(attachments.FooService))
	router.HandleFunc("GET /foo/v1/{id}", newHandleFooV1Get(attachments.FooService))
	httpServer := http.Server{
		Handler: router,
		Addr:    ":8080",
	}
	return httpServer
}

type ServerAttachments struct {
	FooService *foo.Service
}

func newHandleFooV1Post(fooService *foo.Service) http.HandlerFunc {
	return func(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
		ctx := httpRequest.Context()
		httpRequestBody, err := io.ReadAll(httpRequest.Body)
		if err != nil {
			httpResponseWriter.WriteHeader(http.StatusBadRequest)
			return
		}
		var fooRequest foo.CreateBarRequest
		err = json.Unmarshal(httpRequestBody, &fooRequest)
		if err != nil {
			httpResponseWriter.WriteHeader(http.StatusBadRequest)
			return
		}
		fooResponse, err := fooService.CreateBar(ctx, fooRequest)
		if err != nil {
			switch errs.Cause(err).(type) {
			case errs.BadRequest:
				httpResponseWriter.WriteHeader(http.StatusBadRequest)
				return
			default:
				httpResponseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		httpResponseBody, err := json.Marshal(fooResponse)
		if err != nil {
			httpResponseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		httpResponseWriter.Header().Add("Content-Type", "application/json")
		_, err = httpResponseWriter.Write(httpResponseBody)
		if err != nil {
			httpResponseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		// httpResponseWriter.WriteHeader(http.StatusOK) Shown as example, but not needed if there is a successful write to the body.
		// Will actually generate a log message saying a superfluous header write was made or something like that
	}
}

func newHandleFooV1Get(fooService *foo.Service) http.HandlerFunc {
	return func(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
		ctx := httpRequest.Context()
		id := httpRequest.PathValue("id") // The path value here needs to match what was inside the curly braces above, {id}, so "id" here
		fooRequest := foo.GetBarByIdRequest{
			Id: id,
		}
		bar, err := fooService.GetBarById(ctx, fooRequest)
		if err != nil {
			switch errs.Cause(err).(type) {
			case errs.BadRequest:
				httpResponseWriter.WriteHeader(http.StatusBadRequest)
				return
			default:
				httpResponseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		httpResponseBody, err := json.Marshal(bar)
		if err != nil {
			httpResponseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		httpResponseWriter.Header().Add("Content-Type", "application/json")
		_, err = httpResponseWriter.Write(httpResponseBody)
		if err != nil {
			httpResponseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
