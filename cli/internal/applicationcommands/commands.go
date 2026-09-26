package applicationcommands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"spx/internal/auth"
	"spx/internal/platformapi"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{Use: "platform", Short: "SPX Internal Developer Platform CLI"}
	root.AddCommand(newLoginCommand(), newLogoutCommand(), newWhoAmICommand(), newAppCommand())
	return root
}

func newLoginCommand() *cobra.Command {
	return &cobra.Command{Use: "login", Short: "Authenticate with Microsoft Entra ID", RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := auth.ConfigFromEnv()
		if err != nil {
			return err
		}
		_, err = auth.NewManager(config).Login(cmd.Context())
		if err == nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Login complete.")
		}
		return err
	}}
}

func newLogoutCommand() *cobra.Command {
	return &cobra.Command{Use: "logout", Short: "Clear the cached authentication token", RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := auth.ConfigFromEnv()
		if err != nil {
			return err
		}
		if err := auth.NewManager(config).Logout(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
		return nil
	}}
}

func newWhoAmICommand() *cobra.Command {
	return &cobra.Command{Use: "whoami", Short: "Show the authenticated platform user", RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := clientFromEnv(cmd.Context())
		if err != nil {
			return err
		}
		user, err := client.Me(cmd.Context())
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "User: %s\nDisplay name: %s\nTenant: %s\nAuthenticated: yes\n", user.Email, user.DisplayName, user.TenantID)
		return nil
	}}
}

func newAppCommand() *cobra.Command {
	app := &cobra.Command{Use: "app", Short: "Manage platform applications"}
	app.AddCommand(newCreateCommand(), newListCommand(), newDescribeCommand(), newStatusCommand())
	return app
}

func newCreateCommand() *cobra.Command {
	var name, appType, runtime, owningTeam string
	command := &cobra.Command{Use: "create", Short: "Register an application intent", RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := clientFromEnv(cmd.Context())
		if err != nil {
			return err
		}
		application, err := client.CreateApplication(cmd.Context(), platformapi.CreateApplicationRequest{Name: name, Type: appType, Runtime: runtime, OwningTeam: owningTeam})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created %s\nProvisioning: %s\nRequest: %s\n", application.Name, application.ProvisioningRequest.Status, application.ProvisioningRequest.ID)
		return nil
	}}
	command.Flags().StringVar(&name, "name", "", "application name")
	command.Flags().StringVar(&appType, "type", "api", "application type: api, worker, web")
	command.Flags().StringVar(&runtime, "runtime", "", "application runtime")
	command.Flags().StringVar(&owningTeam, "owning-team", "", "owning team identifier")
	_ = command.MarkFlagRequired("name")
	_ = command.MarkFlagRequired("runtime")
	_ = command.MarkFlagRequired("owning-team")
	return command
}

func newListCommand() *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List applications", RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := clientFromEnv(cmd.Context())
		if err != nil {
			return err
		}
		applications, err := client.ListApplications(cmd.Context())
		if err != nil {
			return err
		}
		for _, app := range applications {
			status := "PENDING"
			if app.ProvisioningRequest != nil {
				status = app.ProvisioningRequest.Status
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-8s %-12s %s\n", app.Name, app.Type, app.OwningTeam, status)
		}
		return nil
	}}
}

func newDescribeCommand() *cobra.Command {
	return &cobra.Command{Use: "describe <id>", Args: cobra.ExactArgs(1), Short: "Describe an application", RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromEnv(cmd.Context())
		if err != nil {
			return err
		}
		app, err := client.GetApplication(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\nType: %s\nRuntime: %s\nOwning team: %s\nCreated by: %s\nProvisioning: %s\n", app.Name, app.Type, app.Runtime, app.OwningTeam, app.CreatedBy.Email, app.ProvisioningRequest.Status)
		return nil
	}}
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{Use: "status <request-id>", Args: cobra.ExactArgs(1), Short: "Show a provisioning request", RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromEnv(cmd.Context())
		if err != nil {
			return err
		}
		request, err := client.GetProvisioning(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Request: %s\nStatus: %s\nError: %s\n", request.ID, request.Status, request.Error)
		return nil
	}}
}

func clientFromEnv(ctx context.Context) (*platformapi.Client, error) {
	config, err := auth.ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	manager := auth.NewManager(config)
	source, err := manager.TokenSource(ctx)
	if err != nil {
		return nil, err
	}
	return platformapi.NewClient(config.APIURL, source), nil
}
