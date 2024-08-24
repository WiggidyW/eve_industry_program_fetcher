package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func GetAndWriteMarketOrders(
	accessToken string,
	locationIds []int64,
	regionIds []int32,
) error {
	serializableLocationOrders, err := GetSerializableLocationOrders(
		accessToken,
		locationIds,
		regionIds,
	)
	if err != nil {
		return err
	}
	return serializableLocationOrders.Write()
}

func GetSerializableLocationOrders(
	accessToken string,
	locationIds []int64,
	regionIds []int32,
) (
	serializableLocationOrders SerializableLocationOrders,
	err error,
) {
	regionOrders, structureOrders, err := GetOrders(
		accessToken,
		locationIds,
		regionIds,
	)
	if err != nil {
		return nil, err
	}
	return OrdersToSerializable(regionOrders, structureOrders), nil
}

func GetOrders(
	accessToken string,
	locationIds []int64,
	regionIds []int32,
) (
	regionOrders [][]OrdersRegionEntry,
	structureOrders map[int64][]OrdersStructureEntry,
	err error,
) {
	chnRegion := make(chan GetOrdersResult[OrdersRegionEntry, int32], len(regionIds))
	for _, v := range regionIds {
		go func(v int32) {
			orders, err := GetRegionOrders(accessToken, v)
			chnRegion <- GetOrdersResult[OrdersRegionEntry, int32]{Id: v, Model: orders, Err: err}
		}(v)
	}

	chnStructure := make(chan GetOrdersResult[OrdersStructureEntry, int64], len(locationIds))
	for _, v := range locationIds {
		go func(v int64) {
			orders, err := GetStructureOrders(accessToken, v)
			chnStructure <- GetOrdersResult[OrdersStructureEntry, int64]{Id: v, Model: orders, Err: err}
		}(v)
	}

	regionOrders = make([][]OrdersRegionEntry, 0, len(regionIds))
	for i := 0; i < len(regionIds); i++ {
		pageResult := <-chnRegion
		if pageResult.Err != nil {
			return nil, nil, pageResult.Err
		}
		regionOrders = append(regionOrders, pageResult.Model)
	}

	structureOrders = make(map[int64][]OrdersStructureEntry, len(locationIds))
	for i := 0; i < len(locationIds); i++ {
		pageResult := <-chnStructure
		if pageResult.Err != nil {
			return nil, nil, pageResult.Err
		}
		structureOrders[pageResult.Id] = pageResult.Model
	}

	return regionOrders, structureOrders, nil
}

type GetOrdersResult[E any, ID any] struct {
	Id    ID
	Model []E
	Err   error
}

func GetStructureOrders(
	accessToken string,
	locationId int64,
) (
	orders []OrdersStructureEntry,
	err error,
) {

	chn, pages, _, err := getPages[[]OrdersStructureEntry](
		fmt.Sprintf(
			"https://esi.evetech.net/latest/markets/structures/%d/?datasource=tranquility",
			locationId,
		),
		accessToken,
		func() *[]OrdersStructureEntry {
			orders := make([]OrdersStructureEntry, 0, 1000)
			return &orders
		},
	)
	if err != nil {
		return nil, err
	}

	orders = make([]OrdersStructureEntry, 0, pages*1000)
	for i := 0; i < pages; i++ {
		pageResult := <-chn
		if pageResult.Err != nil {
			return nil, pageResult.Err
		}
		orders = append(orders, pageResult.Model...)
	}

	return orders, nil
}

func GetRegionOrders(
	accessToken string,
	regionId int32,
) (
	orders []OrdersRegionEntry,
	err error,
) {
	chn, pages, _, err := getPages[[]OrdersRegionEntry](
		fmt.Sprintf(
			"https://esi.evetech.net/latest/markets/%d/orders/?datasource=tranquility",
			regionId,
		),
		accessToken,
		func() *[]OrdersRegionEntry {
			orders := make([]OrdersRegionEntry, 0, 1000)
			return &orders
		},
	)
	if err != nil {
		return nil, err
	}

	orders = make([]OrdersRegionEntry, 0, pages*1000)
	for i := 0; i < pages; i++ {
		pageResult := <-chn
		if pageResult.Err != nil {
			return nil, pageResult.Err
		}
		orders = append(orders, pageResult.Model...)
	}

	return orders, nil
}

type OrdersRegionEntry struct {
	LocationId   int64   `json:"location_id"`
	Price        float64 `json:"price"`
	TypeId       int32   `json:"type_id"`
	VolumeRemain int32   `json:"volume_remain"`
}

type OrdersStructureEntry struct {
	IsBuyOrder   bool    `json:"is_buy_order"`
	Price        float64 `json:"price"`
	TypeId       int32   `json:"type_id"`
	VolumeRemain int32   `json:"volume_remain"`
}

type SerializableOrder struct {
	Price  float64 `json:"price"`
	Volume uint64  `json:"volume"`
}

type SerializableTypeOrders struct {
	Orders []SerializableOrder `json:"orders"`
	Total  uint64              `json:"total"`
}

type SerializableOrders map[int32]*SerializableTypeOrders

type SerializableLocationOrders map[int64]SerializableOrders

func (s SerializableLocationOrders) Serialize() ([]byte, error) { return json.Marshal(s) }

func (s SerializableLocationOrders) Write() error {
	data, err := s.Serialize()
	if err != nil {
		return err
	}
	return os.WriteFile("market_orders.json", data, 0644)
}

func OrdersToSerializable(
	regionOrders [][]OrdersRegionEntry,
	structureOrders map[int64][]OrdersStructureEntry,
) SerializableLocationOrders {
	serializableLocationOrders := make(SerializableLocationOrders)
	for _, v := range regionOrders {
		WithRegionOrders(serializableLocationOrders, v)
	}
	for k, v := range structureOrders {
		WithStructureOrders(serializableLocationOrders, v, k)
	}
	return serializableLocationOrders
}

func WithRegionOrders(
	serializableLocationOrders map[int64]SerializableOrders,
	orders []OrdersRegionEntry,
) {
	for _, v := range orders {
		if v.VolumeRemain <= 0 {
			continue
		}

		serializableOrders, ok := serializableLocationOrders[v.LocationId]
		if !ok {
			serializableOrders = SerializableOrders{}
			serializableLocationOrders[v.LocationId] = serializableOrders
		}

		serializableTypeOrders, ok := serializableOrders[v.TypeId]
		if !ok {
			serializableTypeOrders = &SerializableTypeOrders{
				Orders: []SerializableOrder{},
			}
			serializableOrders[v.TypeId] = serializableTypeOrders
		}

		serializableTypeOrders.Orders = append(
			serializableTypeOrders.Orders,
			SerializableOrder{
				Price:  v.Price,
				Volume: uint64(v.VolumeRemain),
			},
		)
		serializableTypeOrders.Total += uint64(v.VolumeRemain)
	}
}

func WithStructureOrders(
	serializableLocationOrders map[int64]SerializableOrders,
	orders []OrdersStructureEntry,
	locationId int64,
) {
	for _, v := range orders {
		if v.IsBuyOrder || v.VolumeRemain <= 0 {
			continue
		}

		serializableOrders, ok := serializableLocationOrders[locationId]
		if !ok {
			serializableOrders = SerializableOrders{}
			serializableLocationOrders[locationId] = serializableOrders
		}

		serializableTypeOrders, ok := serializableOrders[v.TypeId]
		if !ok {
			serializableTypeOrders = &SerializableTypeOrders{
				Orders: []SerializableOrder{},
			}
			serializableOrders[v.TypeId] = serializableTypeOrders
		}

		serializableTypeOrders.Orders = append(
			serializableTypeOrders.Orders,
			SerializableOrder{
				Price:  v.Price,
				Volume: uint64(v.VolumeRemain),
			},
		)
		serializableTypeOrders.Total += uint64(v.VolumeRemain)
	}
}
