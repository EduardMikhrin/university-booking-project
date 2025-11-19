// @title University Booking API
// @version 1.0
// @description Backend API for university booking system
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /internal/server
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
