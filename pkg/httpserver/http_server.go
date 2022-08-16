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

type routeFunc func(*gin.Context, resources.ResourceType, *store.StoreConfig)

func curryRoute(f routeFunc, resourceType resources.ResourceType, storeConfig *store.StoreConfig) gin.HandlerFunc {
	return func(g *gin.Context) {
		f(g, resourceType, storeConfig)
	}
}

func readinessRoute(c *gin.Context) {
	c.String(http.StatusOK, "Ok")
}

func resourcesRoute(c *gin.Context, resourceType resources.ResourceType, storeConfig *store.StoreConfig) {
	if resourceType == resources.ResourceTypeUnknown {
		c.String(http.StatusBadRequest, "Resource type unknown")
		return
	}
	filePath := storeConfig.GetFilePath(resourceType)
	if !storeConfig.FileStoreExists(resourceType) {
		c.String(http.StatusNotFound, fmt.Sprintf("file %s not found", filePath))
		return
	}
	logrus.Debugf("Serving file %s", filePath)
	c.File(filePath)
}

func setupRouter(storeConfig *store.StoreConfig) *gin.Engine {
	router := gin.Default()
	skipLogs := []string{
		"/health",
	}
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, skipLogs...))
	router.Use(gin.Recovery())
	router.GET("/readiness", readinessRoute)
	resourceRoute := router.Group("/k8s/resources")
	{
		for r := resources.ResourceTypeApiResource; r < resources.ResourceTypeUnknown; r++ {
			resourceRoute.GET(r.String(), curryRoute(resourcesRoute, r, storeConfig))
		}
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

func StartHttpServer(ctx context.Context, h *HttpServerConfigCli, storeConfig *store.StoreConfig) (int, error) {
	if h.ListenAddress == "" {
		return 0, nil
	}
	if h.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	listener, err := net.Listen("tcp", h.ListenAddress)
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	router := setupRouter(storeConfig)
	srv := &http.Server{
		Addr:    h.ListenAddress,
		Handler: router,
	}
	go startHttpServer(ctx, listener, srv)
	return port, nil
}
