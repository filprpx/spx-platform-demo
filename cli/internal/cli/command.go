package cli

import (
	"context"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"spx/internal/auth"
	"spx/internal/config"
	"spx/internal/platform"
)

type Dependencies struct {
	LoadConfig func() (config.Config, error)
	NewAuth    func(config.Config) *auth.Manager
	NewAPI     func(string, oauth2.TokenSource) *platform.Client
}

func NewCommand(dependencies Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:           "spx",
		Short:         "SPX internal developer platform CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
	}
	root.AddCommand(
		newLoginCommand(dependencies),
		newLogoutCommand(dependencies),
		newWhoAmICommand(dependencies),
		newApplicationCommand(dependencies),
	)
	return root
}

func authenticatedClient(ctx context.Context, dependencies Dependencies) (*platform.Client, error) {
	config, err := dependencies.LoadConfig()
	if err != nil {
		return nil, err
	}
	manager := dependencies.NewAuth(config)
	source, err := manager.TokenSource(ctx)
	if err != nil {
		return nil, err
	}
	return dependencies.NewAPI(config.APIURL, source), nil
}
