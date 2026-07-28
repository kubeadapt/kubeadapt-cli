package cmd

import (
	"cmp"
	"context"
	"os"
	"strings"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
)

// Completion runs on every TAB press, so it gets a budget an order of
// magnitude tighter than a normal request and a hard cap on candidates: the
// reference org carries 3891 workloads, and a shell menu that large is noise.
var (
	completionTimeout    = 2 * time.Second
	completionMaxResults = 100
)

// completionConfig resolves credentials without the RunContext: cobra answers
// __complete straight from getCompletions and never runs PersistentPreRunE,
// so the context root builds is absent here.
func completionConfig(cmd *cobra.Command) *config.Config {
	if rc := getRunContext(cmd); rc != nil && rc.Config != nil {
		return rc.Config
	}
	cfg, err := config.Load(cfgFile)
	if err != nil {
		cfg = config.Default()
	}
	cfg.APIURL = cmp.Or(apiURL, os.Getenv("KUBEADAPT_API_URL"), cfg.APIURL)
	cfg.APIKey = cmp.Or(apiKey, os.Getenv("KUBEADAPT_API_KEY"), cfg.APIKey)
	return cfg
}

func completionClient(cmd *cobra.Command) (*api.Client, bool) {
	cfg := completionConfig(cmd)
	if cfg == nil || cfg.APIKey == "" {
		return nil, false
	}
	return api.NewClient(cfg.APIURL, cfg.APIKey, api.WithTimeout(completionTimeout)), true
}

// Every failure - unauthenticated, offline, slow, malformed - collapses to the
// same empty NoFileComp result: a TAB press must never print, never hang, and
// never degrade into offering filenames where an ID belongs.
func completeIDs[T any](
	cmd *cobra.Command,
	toComplete string,
	list func(context.Context, *api.Client) ([]T, error),
	describe func(T) (id, name string),
) ([]cobra.Completion, cobra.ShellCompDirective) {
	c, ok := completionClient(cmd)
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	parent := cmd.Context()
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, completionTimeout)
	defer cancel()

	items, err := list(ctx, c)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	out := make([]cobra.Completion, 0, min(len(items), completionMaxResults))
	for _, item := range items {
		id, name := describe(item)
		if !strings.HasPrefix(id, toComplete) {
			continue
		}
		out = append(out, cobra.CompletionWithDesc(id, name))
		if len(out) == completionMaxResults {
			break
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// firstArgOnly stops a detail command from spending a request completing a
// second positional value that its own ExactArgs(1) would reject anyway.
func firstArgOnly(fn cobra.CompletionFunc) cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return fn(cmd, args, toComplete)
	}
}

func staticCompletions(values ...string) cobra.CompletionFunc {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		out := make([]cobra.Completion, 0, len(values))
		for _, v := range values {
			if strings.HasPrefix(v, toComplete) {
				out = append(out, v)
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	}
}

// The error is dropped deliberately: it fires only on a misspelled flag name,
// which the flag's own usage-string tests already catch.
func registerEnumFlag(cmd *cobra.Command, flag string, values ...string) {
	_ = cmd.RegisterFlagCompletionFunc(flag, staticCompletions(values...))
}

func registerClusterIDFlag(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("cluster-id", completeClusterIDs)
}

func completeClusterIDs(cmd *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return completeIDs(cmd, toComplete,
		func(ctx context.Context, c *api.Client) ([]types.Cluster, error) {
			items, _, err := c.ListClusters(ctx, api.ClusterFilter{PagedOpts: completionPage()})
			return items, err
		},
		func(v types.Cluster) (string, string) { return v.ID, v.Metadata.Name },
	)
}

func completeWorkloadUIDs(cmd *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return completeIDs(cmd, toComplete,
		func(ctx context.Context, c *api.Client) ([]types.Workload, error) {
			items, _, err := c.ListWorkloads(ctx, api.WorkloadFilter{PagedOpts: completionPage()})
			return items, err
		},
		func(v types.Workload) (string, string) { return v.ID, v.Metadata.Name },
	)
}

func completeNodeUIDs(cmd *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return completeIDs(cmd, toComplete,
		func(ctx context.Context, c *api.Client) ([]types.Node, error) {
			items, _, err := c.ListNodes(ctx, api.NodeFilter{PagedOpts: completionPage()})
			return items, err
		},
		func(v types.Node) (string, string) { return v.ID, v.Metadata.Name },
	)
}

func completeRecommendationIDs(cmd *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return completeIDs(cmd, toComplete,
		func(ctx context.Context, c *api.Client) ([]types.Recommendation, error) {
			items, _, err := c.ListRecommendations(ctx, api.RecommendationFilter{PagedOpts: completionPage()})
			return items, err
		},
		func(v types.Recommendation) (string, string) { return v.ID, v.Metadata.Title },
	)
}

func completeTeamIDs(cmd *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return completeIDs(cmd, toComplete,
		func(ctx context.Context, c *api.Client) ([]types.Team, error) {
			items, _, err := c.ListTeams(ctx, api.TeamFilter{PagedOpts: completionPage()})
			return items, err
		},
		func(v types.Team) (string, string) { return v.ID, v.Metadata.Name },
	)
}

func completeDepartmentIDs(cmd *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return completeIDs(cmd, toComplete,
		func(ctx context.Context, c *api.Client) ([]types.Department, error) {
			items, _, err := c.ListDepartments(ctx, api.DepartmentFilter{PagedOpts: completionPage()})
			return items, err
		},
		func(v types.Department) (string, string) { return v.ID, v.Metadata.Name },
	)
}

func completionPage() api.PagedOpts {
	return api.PagedOpts{Limit: completionMaxResults}
}
