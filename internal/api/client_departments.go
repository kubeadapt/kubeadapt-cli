package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

type DepartmentFilter struct {
	PagedOpts
	CostModeOpt
	Origins []string // csv
}

type DepartmentGetOpts struct {
	CostModeOpt
}

func (c *Client) ListDepartments(
	ctx context.Context, f DepartmentFilter,
) ([]types.Department, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, f.CostMode)
	setCSV(params, "origin", f.Origins)
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.Department](ctx, c, "/v1/departments", params)
}

func (c *Client) GetDepartment(
	ctx context.Context, deptID string, opts DepartmentGetOpts,
) (*types.Department, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, opts.CostMode)
	return DoEnvelopeGet[*types.Department](ctx, c, "/v1/departments/"+deptID, params)
}
