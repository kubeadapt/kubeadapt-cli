package views

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui/components"
)

type PVsView struct {
	client    *api.Client
	data      []types.PersistentVolumeResponse
	table     *components.Table
	loading   bool
	err       error
	spinner   spinner.Model
	filterBar *components.FilterBar
	keys      tui.KeyMap
}

func NewPVsView(client *api.Client) *PVsView {
	return &PVsView{
		client:    client,
		loading:   true,
		spinner:   components.NewSpinner(),
		filterBar: components.NewFilterBar(),
		keys:      tui.DefaultKeyMap(),
	}
}

func (v *PVsView) Title() string { return "Persistent Volumes" }

func (v *PVsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *PVsView) loadData() tea.Cmd {
	return func() tea.Msg {
		data, err := v.client.GetPersistentVolumes(context.Background(), "", "", "")
		return tui.DataLoadedMsg{View: tui.ViewPVs, Data: data, Err: err}
	}
}

func (v *PVsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewPVs {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				resp := msg.Data.(*types.PersistentVolumeListResponse)
				v.data = resp.PersistentVolumes
				v.buildTable()
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	case tea.MouseMsg:
		if v.table != nil {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				v.table.MoveUp()
			case tea.MouseButtonWheelDown:
				v.table.MoveDown()
			}
		}
	case tea.KeyMsg:
		if v.filterBar.IsActive() {
			if v.filterBar.HandleKey(msg) {
				if v.filterBar.IsActive() {
					if v.table != nil {
						v.table.ApplyFilter(v.filterBar.GetQuery())
					}
				} else {
					if v.table != nil {
						v.table.ClearFilter()
					}
				}
				return v, nil
			}
		}
		if v.table != nil {
			switch {
			case key.Matches(msg, v.keys.Filter):
				if !v.filterBar.IsActive() {
					v.filterBar.Activate()
					return v, nil
				}
			case key.Matches(msg, v.keys.Up):
				v.table.MoveUp()
			case key.Matches(msg, v.keys.Down):
				v.table.MoveDown()
			}
		}
	}
	return v, nil
}

func (v *PVsView) buildTable() {
	headers := []string{"Name", "Cluster", "Namespace", "PVC", "Storage Class", "Capacity", "Type", "$/hr"}
	var rows [][]string
	for _, pv := range v.data {
		rows = append(rows, []string{
			pv.Name,
			output.FormatOptionalString(pv.ClusterName),
			output.FormatOptionalString(pv.Namespace),
			output.FormatOptionalString(pv.PVCName),
			output.FormatOptionalString(pv.StorageClass),
			output.FormatMemoryGB(pv.CapacityGB),
			output.FormatOptionalString(pv.VolumeType),
			output.FormatCostPtr(pv.HourlyCost),
		})
	}
	v.table = components.NewTable(headers, rows)
}

func (v *PVsView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading persistent volumes..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.table == nil {
		return "No persistent volumes found"
	}
	v.table.Width = width
	v.table.ViewportHeight = height - 3

	var parts []string
	if v.filterBar.IsActive() {
		v.filterBar.Width = width
		parts = append(parts, v.filterBar.Render())
		v.table.ViewportHeight--
	}
	parts = append(parts, v.table.Render())
	return strings.Join(parts, "\n")
}
