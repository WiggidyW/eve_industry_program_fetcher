package main

import (
	"encoding/json"
)

func GetAndWriteAdjustedPrices(accessToken string) error {
	adjustedPrices, err := GetAdjustedPrices(accessToken)
	if err != nil {
		return err
	}
	return adjustedPrices.Write()
}

func GetSerializableAdjustedPrices(accessToken string) (
	serializableAdjustedPrices SerializableAdjustedPrices,
	err error,
) {
	adjustedPrices, err := GetAdjustedPrices(accessToken)
	if err != nil {
		return nil, err
	}
	return AdjustedPricesToSerializable(adjustedPrices), nil
}

func GetAdjustedPrices(accessToken string) (
	adjustedPrices []AdjustedPriceEntry,
	err error,
) {
	adjustedPrices = make([]AdjustedPriceEntry, 0)
	_, err = getPage[[]AdjustedPriceEntry](
		"https://esi.evetech.net/latest/markets/prices/?datasource=tranquility",
		accessToken,
		&adjustedPrices,
	)
	if err != nil {
		return nil, err
	}

	return adjustedPrices, nil
}

type AdjustedPriceEntry struct {
	AdjustedPrice float64 `json:"adjusted_price"`
	TypeId        int32   `json:"type_id"`
}

type SerializableAdjustedPrices map[int32]float64

func (s SerializableAdjustedPrices) Serialize() ([]byte, error) { return json.Marshal(s) }

func (s SerializableAdjustedPrices) Write() error {
	data, err := s.Serialize()
	if err != nil {
		return err
	}
	return ioutil.WriteFile("adjusted_prices.json", data, 0644)
}

func AdjustedPricesToSerializable(prices []AdjustedPriceEntry) SerializableAdjustedPrices {
	m := make(map[int32]float64)
	for _, v := range prices {
		m[v.TypeId] = v.AdjustedPrice
	}
	return m
}
