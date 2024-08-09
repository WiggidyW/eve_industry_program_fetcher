package main

import (
	"encoding/json"
)

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

func OrdersToSerializable(
	regionOrders []OrdersRegionEntry,
	structureOrders map[int64][]OrdersStructureEntry,
) SerializableLocationOrders {
	serializableLocationOrders := make(SerializableLocationOrders)
	WithRegionOrders(serializableLocationOrders, regionOrders)
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
		if v.IsBuyOrder {
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
