package main

import (
	"encoding/json"
	"os"
)

func GetAndWriteCostIndices(
	accessToken string,
) error {
	serializableCostIndices, err := GetSerializableCostIndices(accessToken)
	if err != nil {
		return err
	}
	return serializableCostIndices.Write()
}

func GetSerializableCostIndices(
	accessToken string,
) (
	serializableCostIndices SerializableCostIndices,
	err error,
) {
	costIndices, err := GetCostIndices(accessToken)
	if err != nil {
		return nil, err
	}
	return CostIndicesToSerializable(costIndices), nil
}

func GetCostIndices(
	accessToken string,
) (
	costIndices []CostIndicesEntry,
	err error,
) {
	costIndices = make([]CostIndicesEntry, 0)
	_, err = getPage[[]CostIndicesEntry](
		"https://esi.evetech.net/latest/industry/systems/?datasource=tranquility",
		accessToken,
		&costIndices,
	)
	if err != nil {
		return nil, err
	}

	return costIndices, nil
}

type CostIndicesSubEntry struct {
	Activity  string  `json:"activity"`
	CostIndex float64 `json:"cost_index"`
}

type CostIndicesEntry struct {
	CostIndices []CostIndicesSubEntry `json:"cost_indices"`
	SystemId    int32                 `json:"solar_system_id"`
}

type SerializableCostIndicesValue struct {
	Manufacturing float64 `json:"manufacturing"`
	Invention     float64 `json:"invention"`
	Reaction      float64 `json:"reaction"`
	Copy          float64 `json:"copy"`
}

type SerializableCostIndices map[int32]SerializableCostIndicesValue

func (s SerializableCostIndices) Serialize() ([]byte, error) { return json.Marshal(s) }

func (s SerializableCostIndices) Write() error {
	data, err := s.Serialize()
	if err != nil {
		return err
	}
	return os.WriteFile("cost_indices.json", data, 0644)
}

func CostIndicesToSerializable(costIndices []CostIndicesEntry) SerializableCostIndices {
	m := make(map[int32]SerializableCostIndicesValue)
	for _, v := range costIndices {
		value := SerializableCostIndicesValue{}
		for _, subEntry := range v.CostIndices {
			switch subEntry.Activity {
			case "manufacturing":
				value.Manufacturing = subEntry.CostIndex
			case "copying":
				value.Copy = subEntry.CostIndex
			case "invention":
				value.Invention = subEntry.CostIndex
			case "reaction":
				value.Reaction = subEntry.CostIndex
			default:
			}
		}
		m[v.SystemId] = value
	}
	return m
}
