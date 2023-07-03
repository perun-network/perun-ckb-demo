package deployment

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
	"perun.network/perun-ckb-backend/backend"
)

const PFLSMinCapacity = 4100000032

type SUDTInfo struct {
	Script  *types.Script
	CellDep *types.CellDep
}

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

func (m Migration) MakeDeployment(systemScripts SystemScripts) (backend.Deployment, *SUDTInfo, error) {
	pcts := m.CellRecipes[0]
	if pcts.Name != "pcts" {
		return backend.Deployment{}, nil, fmt.Errorf("first cell recipe must be pcts")
	}
	pcls := m.CellRecipes[1]
	if pcls.Name != "pcls" {
		return backend.Deployment{}, nil, fmt.Errorf("second cell recipe must be pcls")
	}
	pfls := m.CellRecipes[2]
	if pfls.Name != "pfls" {
		return backend.Deployment{}, nil, fmt.Errorf("third cell recipe must be pfls")
	}
	sudtInfo, err := m.GetSUDT()
	if err != nil {
		return backend.Deployment{}, nil, err
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
		PCTSCodeHash:    types.HexToHash(pcts.DataHash),
		PCTSHashType:    types.HashTypeData1,
		PCLSCodeHash:    types.HexToHash(pcls.DataHash),
		PCLSHashType:    types.HashTypeData1,
		PFLSCodeHash:    types.HexToHash(pfls.DataHash),
		PFLSHashType:    types.HashTypeData1,
		PFLSMinCapacity: PFLSMinCapacity,
		DefaultLockScript: types.Script{
			CodeHash: systemScripts.Secp256k1Blake160SighashAll.ScriptID.CodeHash,
			HashType: systemScripts.Secp256k1Blake160SighashAll.ScriptID.HashType,
			Args:     []byte{},
		},
		DefaultLockScriptDep: systemScripts.Secp256k1Blake160SighashAll.CellDep,
	}, sudtInfo, nil
}

func (m Migration) GetSUDT() (*SUDTInfo, error) {
	sudt := m.CellRecipes[3]
	if sudt.Name != "sudt" {
		return nil, fmt.Errorf("fourth cell recipe must be sudt")
	}

	sudtScript := types.Script{
		CodeHash: types.HexToHash(sudt.DataHash),
		HashType: types.HashTypeData1,
		Args:     []byte{},
	}
	sudtCellDep := types.CellDep{
		OutPoint: &types.OutPoint{},
		DepType:  types.DepTypeCode,
	}
	return &SUDTInfo{
		Script:  &sudtScript,
		CellDep: &sudtCellDep,
	}, nil
}

func GetDeployment(migrationDir, systemScriptsDir string) (backend.Deployment, *SUDTInfo, error) {
	dir, err := os.ReadDir(migrationDir)
	if err != nil {
		return backend.Deployment{}, nil, err
	}
	if len(dir) != 1 {
		return backend.Deployment{}, nil, fmt.Errorf("migration dir must contain exactly one file")
	}
	migrationName := dir[0].Name()
	migrationFile, err := os.Open(path.Join(migrationDir, migrationName))
	defer migrationFile.Close()
	if err != nil {
		return backend.Deployment{}, nil, err
	}
	migrationData, err := io.ReadAll(migrationFile)
	if err != nil {
		return backend.Deployment{}, nil, err
	}
	var migration Migration
	err = json.Unmarshal(migrationData, &migration)
	if err != nil {
		return backend.Deployment{}, nil, err
	}

	ss, err := GetSystemScripts(systemScriptsDir)
	if err != nil {
		return backend.Deployment{}, nil, err
	}
	return migration.MakeDeployment(ss)
}
