package bootstrap

import (
	"github.com/spf13/cobra"
	"spx/internal/auth"
	"spx/internal/cli"
	"spx/internal/config"
	"spx/internal/platform"
)

func NewCLICommand() *cobra.Command {
	return cli.NewCommand(cli.Dependencies{
		LoadConfig: config.FromEnv,
		NewAuth:    auth.NewManager,
		NewAPI:     platform.NewClient,
	})
}
