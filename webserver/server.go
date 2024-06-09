package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/jmhbh/argo-app-orchestrator/types"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server = http.Handler

func Start(ctx context.Context, params Params) error {
	httpServer := &http.Server{
		Addr:    ":9000", // hardcoded port
		Handler: NewServer(params),
	}

	logger := ctx.Value(LoggerKey{}).(*zap.SugaredLogger)
	// run shutdown observer in a goroutine since httpServer.ListenAndServe() is blocking
	go func() {
		<-ctx.Done()

		logger.Infof("shutting web server")

		// give the server 10 seconds to gracefully shutdown
		shutdownTimeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := httpServer.Shutdown(shutdownTimeout); err != nil {
			logger.Errorf("failed to shutdown web server: %v", err)
		}
	}()

	return httpServer.ListenAndServe()
}

func NewServer(params Params) Server {
	e := echo.New()
	// hello world method for testing
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// create method for using user metadata to create a argo application
	e.POST("/api/v1/create", func(c echo.Context) error {
		// get name and email from body

		jsonBody := make(map[string]interface{})
		err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid body: %s", err.Error()))
		}

		name, ok := jsonBody["name"].(string)
		if !ok {
			return c.String(http.StatusBadRequest, "No name param")
		}
		email, ok := jsonBody["email"].(string)
		if !ok {
			return c.String(http.StatusBadRequest, "No email param")
		}

		userMetadata := UserMetadata{
			Name:  name,
			Email: email,
		}

		// send user metadata to the orchestrator
		params.UserMetadataChan <- userMetadata

		for {
			select {
			case <-params.KickChan:
				return c.String(http.StatusOK, fmt.Sprintf("Modified Argo ApplicationSet to manage deployment for additional user: %s", name))
			case <-time.After(10 * time.Second):
				return c.String(http.StatusRequestTimeout, "Timeout")
			}
		}
	})

	return e
}
