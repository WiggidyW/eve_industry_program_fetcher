package main

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

func CostIndicesToSerializable(costIndices []CostIndicesEntry) SerializableCostIndices {
	m := make(map[int32]SerializableCostIndicesValue)
	for _, v := range costIndices {
		value := SerializableCostIndicesValue{
			Manufacturing: v.CostIndices[0].CostIndex,
			Invention:     v.CostIndices[1].CostIndex,
			Reaction:      v.CostIndices[2].CostIndex,
			Copy:          v.CostIndices[3].CostIndex,
		}
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
