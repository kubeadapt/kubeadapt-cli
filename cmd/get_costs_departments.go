package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getCostsDepartmentsCmd = &cobra.Command{
	Use:   "departments",
	Short: "Show cost breakdown by department",
	Example: `  kubeadapt get costs departments
  kubeadapt get costs departments --cluster-id abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching department costs...", func(ctx context.Context) (*types.DepartmentCostListResponse, error) {
			return client.GetCostsDepartments(ctx, clusterID)
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderDepartmentCosts(resp.Departments, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getCostsDepartmentsCmd)
	costsCmd.AddCommand(getCostsDepartmentsCmd)
}
