package deployment

import (
	"encoding/json"
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
	"io"
	"os"
	"path"
	"perun.network/perun-ckb-backend/backend"
)

const PFLSMinCapacity = 4100000032

type Migration struct {
	CellRecipes []struct {
		Name             string      `json:"name"`
		TxHash           string      `json:"tx_hash"`
		Index            uint32      `json:"index"`
		OccupiedCapacity int64       `json:"occupied_capacity"`
		DataHash         string      `json:"data_hash"`
		TypeId           interface{} `json:"type_id"`
	} `json:"cell_recipes"`
	DepGroupRecipes []interface{} `json:"dep_group_recipes"`
}

func (m Migration) MakeDeployment() (backend.Deployment, error) {
	pcts := m.CellRecipes[0]
	if pcts.Name != "pcts" {
		return backend.Deployment{}, fmt.Errorf("first cell recipe must be pcts")
	}
	pcls := m.CellRecipes[1]
	if pcls.Name != "pcls" {
		return backend.Deployment{}, fmt.Errorf("second cell recipe must be pcls")
	}
	pfls := m.CellRecipes[2]
	if pfls.Name != "pfls" {
		return backend.Deployment{}, fmt.Errorf("third cell recipe must be pfls")
	}
	sudt := m.CellRecipes[3]
	if sudt.Name != "sudt" {
		return backend.Deployment{}, fmt.Errorf("fourth cell recipe must be sudt")
	}

	return backend.Deployment{
		Network: types.NetworkTest,
		PCTSDep: types.CellDep{
			OutPoint: &types.OutPoint{
				TxHash: types.HexToHash(pcts.TxHash),
				Index:  m.CellRecipes[0].Index,
			},
			DepType: types.DepTypeCode,
		},
		PCLSDep: types.CellDep{
			OutPoint: &types.OutPoint{
				TxHash: types.HexToHash(pcls.TxHash),
				Index:  m.CellRecipes[0].Index,
			},
			DepType: types.DepTypeCode,
		},
		PFLSDep: types.CellDep{
			OutPoint: &types.OutPoint{
				TxHash: types.HexToHash(pfls.TxHash),
				Index:  m.CellRecipes[0].Index,
			},
			DepType: types.DepTypeCode,
		},
		PCTSCodeHash:      types.HexToHash(pcts.DataHash),
		PCTSHashType:      types.HashTypeData,
		PCLSCodeHash:      types.HexToHash(pcls.DataHash),
		PCLSHashType:      types.HashTypeData,
		PFLSCodeHash:      types.HexToHash(pfls.DataHash),
		PFLSHashType:      types.HashTypeData,
		PFLSMinCapacity:   PFLSMinCapacity,
		DefaultLockScript: types.Script{},
		DefaultLockScriptDep: types.CellDep{ // TODO: These make no sense
			OutPoint: &types.OutPoint{
				TxHash: types.HexToHash(pcls.TxHash),
				Index:  m.CellRecipes[0].Index,
			},
			DepType: types.DepTypeCode,
		},
	}, nil
}

func GetDeployment(migrationDir string) (backend.Deployment, error) {
	dir, err := os.ReadDir(migrationDir)
	if err != nil {
		return backend.Deployment{}, err
	}
	if len(dir) != 1 {
		return backend.Deployment{}, fmt.Errorf("migration dir must contain exactly one file")
	}
	migrationName := dir[0].Name()
	migrationFile, err := os.Open(path.Join(migrationDir, migrationName))
	defer migrationFile.Close()
	if err != nil {
		return backend.Deployment{}, err
	}
	migrationData, err := io.ReadAll(migrationFile)
	if err != nil {
		return backend.Deployment{}, err
	}
	var migration Migration
	err = json.Unmarshal(migrationData, &migration)
	if err != nil {
		return backend.Deployment{}, err
	}
	return migration.MakeDeployment()
}
