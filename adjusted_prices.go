package main

import (
	"encoding/json"
)

type AdjustedPriceEntry struct {
	AdjustedPrice float64 `json:"adjusted_price"`
	TypeId        int32   `json:"type_id"`
}

type SerializableAdjustedPrices map[int32]float64

func (s SerializableAdjustedPrices) Serialize() ([]byte, error) { return json.Marshal(s) }

func AdjustedPricesToSerializable(prices []AdjustedPriceEntry) SerializableAdjustedPrices {
	m := make(map[int32]float64)
	for _, v := range prices {
		m[v.TypeId] = v.AdjustedPrice
	}
	return m
}
