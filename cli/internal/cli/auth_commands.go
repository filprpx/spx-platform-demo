package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLoginCommand(dependencies Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Microsoft Entra ID",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := dependencies.LoadConfig()
			if err != nil {
				return err
			}
			if _, err := dependencies.NewAuth(config).Login(cmd.Context()); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Login complete.")
			return nil
		},
	}
}

func newLogoutCommand(dependencies Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear the cached authentication token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := dependencies.LoadConfig()
			if err != nil {
				return err
			}
			if err := dependencies.NewAuth(config).Logout(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
			return nil
		},
	}
}

func newWhoAmICommand(dependencies Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the authenticated SPX user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := authenticatedClient(cmd.Context(), dependencies)
			if err != nil {
				return err
			}
			user, err := client.Me(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "User: %s\nDisplay name: %s\nTenant: %s\nAuthenticated: yes\n", user.Email, user.DisplayName, user.TenantID)
			return nil
		},
	}
}
