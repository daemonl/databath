package databath

import (
	"database/sql"
	"fmt"
	"strings"
)

type DeleteCheckResult struct {
	ToExecute          []string                                 `json:"-"`
	WillBeDeleted      map[string][]uint64                      `json:"willBeDeleted"`
	PreventingDeletion map[string][]uint64                      `json:"preventsDeletion"`
	Children           map[string]map[string]*DeleteCheckResult `json:"children"`
	Prevents           bool                                     `json:"preventsDeletion"`
}

func (dcr *DeleteCheckResult) AddChild(childDcr *DeleteCheckResult, childCollection string, childId uint64) {
	child, ok := dcr.Children[childCollection]
	if !ok {
		child = make(map[string]*DeleteCheckResult)
		dcr.Children[childCollection] = child
	}

	childIdStr := fmt.Sprintf("%d", childId)

	child[childIdStr] = childDcr

	if childDcr.Prevents {
		dcr.Prevents = true
	}
}

func (dcr *DeleteCheckResult) GetIssues() []string {
	strs := make([]string, 0, 0)
	for collectionName, ids := range dcr.PreventingDeletion {
		idStrings := make([]string, len(ids), len(ids))
		for i, idInt := range ids {
			idStrings[i] = fmt.Sprintf("%d", idInt)
		}
		strs = append(strs, fmt.Sprintf("%s: %s", collectionName, strings.Join(idStrings, ", ")))
	}
	for collectionName, children := range dcr.Children {
		for id, child := range children {
			for _, str := range child.GetIssues() {
				strs = append(strs, collectionName+"["+id+"]."+str)
			}
		}
	}
	return strs
}

func (dcr *DeleteCheckResult) ExecuteRecursive(db *sql.DB) error {

	for _, children := range dcr.Children {
		for _, child := range children {
			err := child.ExecuteRecursive(db)
			if err != nil {
				return err
			}
		}
	}

	for _, sql := range dcr.ToExecute {
		_, err := db.Exec(sql)
		if err != nil {
			return err
		}
	}

	return nil
}
