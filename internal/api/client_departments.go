package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// DepartmentFilter narrows the result set of ListDepartments.
type DepartmentFilter struct {
	PagedOpts
	CostModeOpt
	Origins []string // csv
}

// DepartmentGetOpts captures optional query parameters accepted by
// GET /v1/departments/{dept_id}.
type DepartmentGetOpts struct {
	CostModeOpt
}

// ListDepartments lists departments via GET /v1/departments.
func (c *Client) ListDepartments(
	ctx context.Context, f DepartmentFilter,
) ([]types.Department, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, f.CostMode)
	setCSV(params, "origin", f.Origins)
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.Department](ctx, c, "/v1/departments", params)
}

// GetDepartment fetches a department by ID via GET /v1/departments/{dept_id}.
func (c *Client) GetDepartment(
	ctx context.Context, deptID string, opts DepartmentGetOpts,
) (*types.Department, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, opts.CostMode)
	return DoEnvelopeGet[*types.Department](ctx, c, "/v1/departments/"+deptID, params)
}
