package cmd

import (
	"github.com/EduardMikhrin/university-booking-project/cmd/service/migrate"
	"github.com/EduardMikhrin/university-booking-project/cmd/service/run"
	"github.com/EduardMikhrin/university-booking-project/cmd/utils"
	"github.com/spf13/cobra"
)

func init() {
	registerServiceCommands(Cmd)
	utils.RegisterConfigFlag(Cmd)

}

func registerServiceCommands(cmd *cobra.Command) {
	cmd.AddCommand(migrate.Cmd)
	cmd.AddCommand(run.Cmd)
}

var Cmd = &cobra.Command{
	Use:   "service",
	Short: "Command for running service operations",
}
