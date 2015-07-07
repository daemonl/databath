package model

type DynamicFunction struct {
	Filename   string   `json:"filename"`
	Parameters []string `json:"parameters"`
	Access     []uint64 `json:"access"`
}
