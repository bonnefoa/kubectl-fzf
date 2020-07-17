package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

type HttpServer struct {
	resourceChan map[string]chan string
	httpPort     int
	httpsPort    int
}

func NewHttpServer(resourceChan map[string]chan string, httpPort int, httpsPort int) HttpServer {
	s := HttpServer{}
	s.resourceChan = resourceChan
	s.httpPort = httpPort
	s.httpsPort = httpsPort
	return s
}

func (h *HttpServer) resourceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	outputType := vars["outputType"]
	resourceName := vars["resourceName"]

	glog.Infof("Received %s request for %s", outputType, resourceName)

	if outputType != "resource" && outputType != "header" {
		glog.Infof("Unknown output type %s", outputType)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unknown output type %s", outputType)
		return
	}

	ch, ok := h.resourceChan[resourceName]
	if !ok {
		glog.Infof("Unknown resource type %s", resourceName)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unknown resource %s", resourceName)
		return
	}

	ch <- outputType

	w.WriteHeader(http.StatusOK)
	output := <-ch
	fmt.Fprintf(w, output)
}

func (h *HttpServer) Start(ctx context.Context) {
	r := mux.NewRouter()
	r.HandleFunc("/{outputType}/{resourceName}", h.resourceHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", h.httpPort),
		Handler: r,
	}

	httpsServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", h.httpsPort),
		Handler: r,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			glog.Error(err)
		}
	}()

	// TODO Add key + cert
	go func() {
		if err := httpsServer.ListenAndServe(); err != nil {
			glog.Error(err)
		}
	}()

	<-ctx.Done()
	httpServer.Shutdown(ctx)
	httpsServer.Shutdown(ctx)
}
