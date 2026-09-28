package cli

import (
	"fmt"
	"io"
	"spx/internal/platform"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	labelStyle   = lipgloss.NewStyle().Bold(true)
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
)

func writeTitle(out io.Writer, title string) {
	fmt.Fprintln(out, titleStyle.Render(title))
}

func writeLabel(out io.Writer, label, value string) {
	fmt.Fprintf(out, "%s %s\n", labelStyle.Render(label+":"), value)
}

func writeApplicationList(out io.Writer, applications []platform.Application) {
	if len(applications) == 0 {
		fmt.Fprintln(out, "No applications found.")
		return
	}

	rows := make([]table.Row, 0, len(applications))
	statusWidth := lipgloss.Width("STATUS")
	for _, application := range applications {
		status := "PENDING"
		if application.ProvisioningRequest != nil && application.ProvisioningRequest.Status != "" {
			status = application.ProvisioningRequest.Status
		}
		if width := lipgloss.Width(status); width > statusWidth {
			statusWidth = width
		}
		rows = append(rows, table.Row{
			application.Name,
			application.ComputeSize,
			fmt.Sprintf("%d", application.ContainerPort),
			application.OwningTeam,
			application.Ingress,
			fmt.Sprintf("%d-%d", application.MinReplicas, application.MaxReplicas),
			statusStyle.Render(status),
		})
	}
	statusWidth += 2 // Keep the styled status cell from being clipped at its edges.

	styles := table.DefaultStyles()
	styles.Header = styles.Header.Bold(true).Foreground(lipgloss.Color("62"))
	appTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "NAME", Width: 24},
			{Title: "SIZE", Width: 10},
			{Title: "PORT", Width: 8},
			{Title: "OWNING TEAM", Width: 24},
			{Title: "INGRESS", Width: 10},
			{Title: "REPLICAS", Width: 10},
			{Title: "STATUS", Width: statusWidth},
		}),
		table.WithRows(rows),
		table.WithStyles(styles),
		table.WithFocused(false),
		table.WithWidth(24+10+8+24+10+10+statusWidth+14),
		table.WithHeight(len(rows)+1),
	)
	fmt.Fprintln(out, appTable.View())
}

func provisioningSummary(request *platform.ProvisioningRequest) (id, status string) {
	if request == nil {
		return "-", "not started"
	}
	if request.ID == "" {
		id = "-"
	} else {
		id = request.ID
	}
	if request.Status == "" {
		status = "unknown"
	} else {
		status = request.Status
	}
	return id, status
}

func writeApplicationDetails(out io.Writer, application platform.Application) {
	requestID, requestStatus := provisioningSummary(application.ProvisioningRequest)
	writeTitle(out, "Application: "+application.Name)
	writeLabel(out, "ID", application.ID)
	writeLabel(out, "Owning team", application.OwningTeam)
	writeLabel(out, "Compute size", application.ComputeSize)
	writeLabel(out, "Container port", fmt.Sprintf("%d", application.ContainerPort))
	writeLabel(out, "Ingress", application.Ingress)
	writeLabel(out, "Replicas", fmt.Sprintf("%d-%d", application.MinReplicas, application.MaxReplicas))
	writeLabel(out, "Created by", application.CreatedBy.Email)
	writeLabel(out, "Provisioning request", requestID)
	writeLabel(out, "Provisioning status", requestStatus)
}
