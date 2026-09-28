package cli

import (
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	var (
		interactive   bool
		name          string
		owningTeam    string
		computeSize   string
		containerPort int
		ingress       string
		minReplicas   int
		maxReplicas   int
	)
	command := &cobra.Command{
		Use:   "create",
		Short: "Register an application intent",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("application name is provided with --name; positional arguments are not supported")
			}
			request := platform.CreateApplicationRequest{
				Name: name, OwningTeam: owningTeam, ComputeSize: computeSize,
				ContainerPort: containerPort, Ingress: ingress,
				MinReplicas: minReplicas, MaxReplicas: maxReplicas,
			}
			if interactive || !createFlagsChanged(cmd) {
				if !interactiveTerminal(cmd.InOrStdin(), cmd.OutOrStdout()) {
					return errors.New("interactive application creation requires a terminal; provide create flags or run with -i in a terminal")
				}
				var err error
				request, err = runCreateWizard(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), request)
				if err != nil {
					return err
				}
			}
			if err := validateCreateRequest(request); err != nil {
				return err
			}
			client, err := authenticatedClient(cmd.Context(), dependencies)
			if err != nil {
				return err
			}
			application, err := client.CreateApplication(cmd.Context(), request)
			if err != nil {
				return err
			}
			writeTitle(cmd.OutOrStdout(), "Application created successfully")
			writeApplicationDetails(cmd.OutOrStdout(), application)
			return nil
		},
	}
	command.Flags().BoolVarP(&interactive, "interactive", "i", false, "run the interactive application wizard")
	command.Flags().StringVar(&name, "name", "", "application name")
	command.Flags().StringVar(&owningTeam, "owning-team", "", "owning team identifier")
	command.Flags().StringVar(&computeSize, "compute-size", "", "compute size: small or medium")
	command.Flags().IntVar(&containerPort, "container-port", 8080, "container port")
	command.Flags().StringVar(&ingress, "ingress", "external", "network visibility: external or internal")
	command.Flags().IntVar(&minReplicas, "min-replicas", 0, "minimum replicas")
	command.Flags().IntVar(&maxReplicas, "max-replicas", 1, "maximum replicas")
	return command
}

func createFlagsChanged(command *cobra.Command) bool {
	for _, name := range []string{"name", "owning-team", "compute-size", "container-port", "ingress", "min-replicas", "max-replicas"} {
		if command.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func interactiveTerminal(in io.Reader, out io.Writer) bool {
	inFile, inOK := in.(*os.File)
	outFile, outOK := out.(*os.File)
	return inOK && outOK && term.IsTerminal(int(inFile.Fd())) && term.IsTerminal(int(outFile.Fd()))
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
