package cli

import (
	"github.com/spf13/cobra"
	"spx/internal/platform"
)

func newApplicationCommand(dependencies Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "app",
		Short: "Manage SPX applications",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(
		newCreateCommand(dependencies),
		newListCommand(dependencies),
		newDescribeCommand(dependencies),
		newStatusCommand(dependencies),
	)
	return command
}

func newCreateCommand(dependencies Dependencies) *cobra.Command {
	var appType, runtime, owningTeam string
	command := &cobra.Command{
		Use:   "create <name>",
		Short: "Register an application intent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := authenticatedClient(cmd.Context(), dependencies)
			if err != nil {
				return err
			}
			application, err := client.CreateApplication(cmd.Context(), platform.CreateApplicationRequest{Name: args[0], Type: appType, Runtime: runtime, OwningTeam: owningTeam})
			if err != nil {
				return err
			}
			writeTitle(cmd.OutOrStdout(), "Application created successfully")
			writeApplicationDetails(cmd.OutOrStdout(), application)
			return nil
		},
	}
	command.Flags().StringVar(&appType, "type", "api", "application type: api, worker, web")
	command.Flags().StringVar(&runtime, "runtime", "", "application runtime")
	command.Flags().StringVar(&owningTeam, "owning-team", "", "owning team identifier")
	_ = command.MarkFlagRequired("runtime")
	_ = command.MarkFlagRequired("owning-team")
	return command
}

func newListCommand(dependencies Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List applications",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := authenticatedClient(cmd.Context(), dependencies)
			if err != nil {
				return err
			}
			applications, err := client.ListApplications(cmd.Context())
			if err != nil {
				return err
			}
			writeApplicationList(cmd.OutOrStdout(), applications)
			return nil
		},
	}
}

func newDescribeCommand(dependencies Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe an application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := authenticatedClient(cmd.Context(), dependencies)
			if err != nil {
				return err
			}
			application, err := client.GetApplication(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			writeApplicationDetails(cmd.OutOrStdout(), application)
			return nil
		},
	}
}

func newStatusCommand(dependencies Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "status <request-id>",
		Short: "Show a provisioning request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := authenticatedClient(cmd.Context(), dependencies)
			if err != nil {
				return err
			}
			request, err := client.GetProvisioning(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			writeTitle(cmd.OutOrStdout(), "Provisioning request")
			writeLabel(cmd.OutOrStdout(), "ID", request.ID)
			writeLabel(cmd.OutOrStdout(), "Application", request.Application)
			writeLabel(cmd.OutOrStdout(), "Status", statusStyle.Render(request.Status))
			if request.Error != "" {
				writeLabel(cmd.OutOrStdout(), "Error", request.Error)
			}
			return nil
		},
	}
}
