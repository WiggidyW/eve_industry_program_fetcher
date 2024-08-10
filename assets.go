package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

func GetAndWriteAssets(
	accessToken string,
	corporationId int32,
) error {
	serializableLocationOutAssets, err := GetSerializableLocationOutAssets(
		accessToken,
		corporationId,
	)
	if err != nil {
		return err
	}
	return serializableLocationOutAssets.Write()
}

func GetSerializableLocationOutAssets(
	accessToken string,
	corporationId int32,
) (
	serializableLocationOutAssets SerializableLocationOutAssets,
	err error,
) {
	assets, blueprints, err := GetAssetsAndBlueprints(accessToken, corporationId)
	if err != nil {
		return nil, err
	}
	return AssetsToSerializable(assets, blueprints), nil
}

func GetAssetsAndBlueprints(
	accessToken string,
	corporationId int32,
) (
	assets []AssetsEntry,
	blueprints []BlueprintsEntry,
	err error,
) {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	var assetsErr error
	go func() {
		assets, assetsErr = GetAssets(accessToken, corporationId)
		wg.Done()
	}()

	blueprints, blueprintsErr := GetBlueprints(accessToken, corporationId)
	if blueprintsErr != nil {
		return nil, nil, blueprintsErr
	}

	wg.Wait()
	if assetsErr != nil {
		return nil, nil, assetsErr
	}

	return assets, blueprints, nil
}

func GetBlueprints(
	accessToken string,
	corporationId int32,
) (
	blueprints []BlueprintsEntry,
	err error,
) {
	chn, pages, _, err := getPages[[]BlueprintsEntry](
		fmt.Sprintf(
			"https://esi.evetech.net/latest/corporations/%d/blueprints/?datasource=tranquility",
			corporationId,
		),
		accessToken,
		func() *[]BlueprintsEntry {
			blueprints := make([]BlueprintsEntry, 0, 1000)
			return &blueprints
		},
	)
	if err != nil {
		return nil, err
	}

	blueprints = make([]BlueprintsEntry, 0, pages*1000)
	for i := 0; i < pages; i++ {
		pageResult := <-chn
		if pageResult.Err != nil {
			return nil, pageResult.Err
		}
		blueprints = append(blueprints, pageResult.Model...)
	}

	return blueprints, nil
}

func GetAssets(
	accessToken string,
	corporationId int32,
) (
	assets []AssetsEntry,
	err error,
) {
	chn, pages, _, err := getPages[[]AssetsEntry](
		fmt.Sprintf(
			"https://esi.evetech.net/latest/corporations/%d/assets/?datasource=tranquility",
			corporationId,
		),
		accessToken,
		func() *[]AssetsEntry {
			assets := make([]AssetsEntry, 0, 1000)
			return &assets
		},
	)
	if err != nil {
		return nil, err
	}

	assets = make([]AssetsEntry, 0, pages*1000)
	for i := 0; i < pages; i++ {
		pageResult := <-chn
		if pageResult.Err != nil {
			return nil, pageResult.Err
		}
		assets = append(assets, pageResult.Model...)
	}

	return assets, nil
}

type HasItemId interface {
	GetItemId() int64
}

type AssetsEntry struct {
	ItemId     int64 `json:"item_id"`
	LocationId int64 `json:"location_id"`
	Quantity   int64 `json:"quantity"`
	TypeId     int32 `json:"type_id"`
}

func (a AssetsEntry) GetItemId() int64 { return a.ItemId }

type BlueprintsEntry struct {
	ItemId             int64 `json:"item_id"`
	Runs               int32 `json:"runs"`
	MaterialEfficiency int32 `json:"material_efficiency"`
	TimeEfficiency     int32 `json:"time_efficiency"`
}

func (b BlueprintsEntry) GetItemId() int64 { return b.ItemId }

func ToItemIdMap[T HasItemId](slice []T) map[int64]T {
	m := make(map[int64]T)
	for _, v := range slice {
		m[v.GetItemId()] = v
	}
	return m
}

type OutAsset struct {
	TypeId             int32 `json:"type_id"`
	Runs               int32 `json:"runs"`
	MaterialEfficiency int32 `json:"me"`
	TimeEfficiency     int32 `json:"te"`
}

type LocationOutAssets map[int64]map[OutAsset]int64

type SerializableOutAsset struct {
	OutAsset
	Quantity int64 `json:"quantity"`
}

type SerializableLocationOutAssets map[int64][]SerializableOutAsset

func (s SerializableLocationOutAssets) Serialize() ([]byte, error) { return json.Marshal(s) }

func (s SerializableLocationOutAssets) Write() error {
	data, err := s.Serialize()
	if err != nil {
		return err
	}
	return ioutil.WriteFile("assets.json", data, 0644)
}

func AssetsToSerializable(assets []AssetsEntry, blueprints []BlueprintsEntry) SerializableLocationOutAssets {
	locationOutAssets := make(LocationOutAssets)
	assetsMap := ToItemIdMap(assets)
	blueprintsMap := ToItemIdMap(blueprints)

	for _, asset := range assets {
		locationId := asset.LocationId

		outAsset := OutAsset{
			TypeId: asset.TypeId,
		}

		if blueprint, ok := blueprintsMap[asset.ItemId]; ok {
			outAsset.Runs = blueprint.Runs
			outAsset.MaterialEfficiency = blueprint.MaterialEfficiency
			outAsset.TimeEfficiency = blueprint.TimeEfficiency
		}

		for {
			parentAsset, ok := assetsMap[locationId]
			if ok {
				locationId = parentAsset.LocationId
			} else {
				break
			}
		}

		if _, ok := locationOutAssets[locationId]; !ok {
			locationOutAssets[locationId] = make(map[OutAsset]int64)
		}

		// this will create a new entry if it doesn't exist
		locationOutAssets[locationId][outAsset] += asset.Quantity
	}

	serializableLocationOutAssets := make(SerializableLocationOutAssets)
	for locationId, outAssets := range locationOutAssets {
		serializableOutAssets := make([]SerializableOutAsset, 0, len(outAssets))
		for outAsset, quantity := range outAssets {
			serializableOutAssets = append(serializableOutAssets, SerializableOutAsset{
				OutAsset: outAsset,
				Quantity: quantity,
			})
		}
		serializableLocationOutAssets[locationId] = serializableOutAssets
	}

	return serializableLocationOutAssets
}
