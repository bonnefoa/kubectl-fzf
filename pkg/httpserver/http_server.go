package httpserver

import (
	"context"
	"fmt"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/k8s/store"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type FzfHttpServer struct {
	Port        int
	ResourceHit int
	//LastModifiedHit int

	stores      []*store.Store
	storeConfig *store.StoreConfig
}

type routeResourceFunc func(*gin.Context, resources.ResourceType)

func curryResourceRoute(f routeResourceFunc, resourceType resources.ResourceType) gin.HandlerFunc {
	return func(g *gin.Context) {
		f(g, resourceType)
	}
}

func (f *FzfHttpServer) readinessRoute(c *gin.Context) {
	c.String(http.StatusOK, "Ok")
}

func (f *FzfHttpServer) statsRoute(c *gin.Context) {
	stats := store.GetStatsFromStores(f.stores)
	logrus.Debugf("Sending stats: %v", stats)
	c.JSON(http.StatusOK, stats)
}

func (f *FzfHttpServer) resourcesRoute(c *gin.Context, resourceType resources.ResourceType) {
	if c.Request.Method == "GET" {
		f.ResourceHit++
	}
	if resourceType == resources.ResourceTypeUnknown {
		c.String(http.StatusBadRequest, "Resource type unknown")
		return
	}
	if !f.storeConfig.FileStoreExists(resourceType) {
		c.String(http.StatusNotFound, fmt.Sprintf("resource file for %s not found", resourceType))
		return
	}
	filePath := f.storeConfig.GetFilePath(resourceType)
	logrus.Debugf("Serving file %s", filePath)
	c.File(filePath)
}

func (f *FzfHttpServer) setupRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	skipLogs := []string{
		"/health",
	}
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, skipLogs...))
	router.Use(gin.Recovery())
	router.GET("/readiness", f.readinessRoute)
	router.GET("/stats", f.statsRoute)

	resourceRoute := router.Group("/k8s/resources")
	for r := resources.ResourceTypeApiResource; r < resources.ResourceTypeUnknown; r++ {
		resourceRoute.GET(r.String(), curryResourceRoute(f.resourcesRoute, r))
		resourceRoute.HEAD(r.String(), curryResourceRoute(f.resourcesRoute, r))
	}

	return router
}

func startHttpServer(ctx context.Context, listener net.Listener, srv *http.Server) {
	go func() {
		logrus.Infof("Starting http server on %s", srv.Addr)
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("error listening: %s", err)
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %s", err)
	}
	logrus.Info("Exiting http server")
}

func StartHttpServer(ctx context.Context, h *HttpServerConfigCli, storeConfig *store.StoreConfig, stores []*store.Store) (*FzfHttpServer, error) {
	if h.ListenAddress == "" {
		return nil, nil
	}
	if h.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	listener, err := net.Listen("tcp", h.ListenAddress)
	if err != nil {
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	f := FzfHttpServer{
		Port:        port,
		stores:      stores,
		storeConfig: storeConfig,
	}
	router := f.setupRouter()
	srv := &http.Server{
		Addr:    h.ListenAddress,
		Handler: router,
	}
	go startHttpServer(ctx, listener, srv)
	return &f, nil
}
