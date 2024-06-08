package main

import (
	"context"
	"github.com/jmhbh/argo-app-orchestrator/app_orchestrator"
	. "github.com/jmhbh/argo-app-orchestrator/bootstrap"
	. "github.com/jmhbh/argo-app-orchestrator/types"
	"github.com/jmhbh/argo-app-orchestrator/webserver"
	"go.uber.org/zap"
)

func main() {
	// Setup logger and inject it into the context
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	ctx = context.WithValue(ctx, LoggerKey{}, sugar)

	// list startFuncs
	startFuncs := map[string]StartFunc{
		"webserver":        webserver.Start,
		"app-orchestrator": app_orchestrator.Start,
	}

	// init channels to share data between components
	params := Params{
		UserMetadataChan: make(chan UserMetadata),
		KickChan:         make(chan struct{}),
	}

	// start multiple components concurrently in separate goroutines
	// so far we have the webserver component, and the argo-orchestrator component
	if err := StartMulti(ctx, startFuncs, params); err != nil {
		sugar.Fatalf("error starting services: %v", err)
	}
	sugar.Infof("exiting...")
}
