package models

type Integration struct {
	Uuid          string `json:"uuid"`
	ApplicationID int64  `json:"applicationId"`
	DatasetNodeID string `json:"datasetId"`
}
