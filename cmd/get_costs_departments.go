package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getCostsDepartmentsCmd = &cobra.Command{
	Use:   "departments",
	Short: "Show cost breakdown by department",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		resp, err := client.GetCostsDepartments(cmd.Context(), clusterID)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderDepartmentCosts(resp.Departments, noColor)
		})
	},
}

func init() {
	getCostsDepartmentsCmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	costsCmd.AddCommand(getCostsDepartmentsCmd)
}
