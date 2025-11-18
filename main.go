package main

import (
	"github.com/EduardMikhrin/university-booking-project/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "booking-svc",
		Short: "Booking service application",
	}

	rootCmd.AddCommand(cmd.Cmd)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
