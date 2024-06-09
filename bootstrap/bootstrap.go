package bootstrap

import (
	"context"
	. "github.com/jmhbh/argo-app-orchestrator/types"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type StartFunc func(ctx context.Context, params Params) error

func StartMulti(ctx context.Context, startFuncs map[string]StartFunc, params Params) error {
	eg := errgroup.Group{}
	for name, start := range startFuncs {
		start := start
		name := name

		logger := ctx.Value(LoggerKey{}).(*zap.SugaredLogger)
		eg.Go(func() error {
			logger.Infof("starting component: %s", name)
			err := start(ctx, params)
			if err != nil {
				logger.Errorf("start func: %s encountered err: %s", name, err.Error())
			}
			return err
		})
	}
	return eg.Wait() // returns the first error encountered
}
