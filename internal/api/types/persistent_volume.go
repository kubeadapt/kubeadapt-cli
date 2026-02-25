package types

// PersistentVolumeResponse represents a single persistent volume.
type PersistentVolumeResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	ClusterID    string   `json:"cluster_id"`
	ClusterName  *string  `json:"cluster_name"`
	Namespace    *string  `json:"namespace"`
	PVCName      *string  `json:"pvc_name"`
	StorageClass *string  `json:"storage_class"`
	CapacityGB   float64  `json:"capacity_gb"`
	AccessModes  []string `json:"access_modes"`
	VolumeType   *string  `json:"volume_type"`
	Zone         *string  `json:"zone"`
	HourlyCost   *float64 `json:"hourly_cost"`
}

// PersistentVolumeListResponse is a list of persistent volumes.
type PersistentVolumeListResponse struct {
	PersistentVolumes []PersistentVolumeResponse `json:"persistent_volumes"`
	Total             int                        `json:"total"`
	Summary           PVSummary                  `json:"summary"`
}
