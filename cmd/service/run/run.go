package run

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/EduardMikhrin/university-booking-project/cmd/utils"
	"github.com/EduardMikhrin/university-booking-project/internal/config"
	"github.com/EduardMikhrin/university-booking-project/internal/data/postgres"
	"github.com/EduardMikhrin/university-booking-project/internal/server"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	utils.RegisterConfigFlag(Cmd)

}

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := utils.ConfigFromFlags(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to get config from flags")
		}

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
		defer cancel()

		err = runService(ctx, cfg)

		return errors.Wrap(err, "failed to run relayer service")
	},
}

func runService(ctx context.Context, cfg config.Config) error {
	wg := new(sync.WaitGroup)
	eg, ctx := errgroup.WithContext(ctx)
	sqlxDB := sqlx.NewDb(cfg.DB().RawDB(), "postgres")
	db := postgres.NewMaster(sqlxDB)

	wg.Add(1)
	eg.Go(func() error {
		server := server.NewServer(cfg.Log(), db, cfg.Cache(), cfg.ApiHttpListener(), cfg.JWT())
		return server.Run(ctx)
	})

	err := eg.Wait()
	wg.Wait()

	return err
}
