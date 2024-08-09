package main

import (
	"flag"
	"fmt"
	"os"
	"log"
)

func main() {
	get_adjusted_prices := flag.Bool("adjusted_prices", false, "Get adjusted prices")
	get_cost_indexes := flag.Bool("cost_indexes", false, "Get cost indexes")
	get_market_orders := flag.Bool("market_orders", false, "Get market orders")
	get_assets := flag.Bool("assets", false, "Get assets")
	flag.Parse()

	config, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	accessToken, _, err := authenticate(
		config.ClientId,
		config.ClientSecret,
		config.RefreshToken,
	)
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	results := make(chan error, 4)

	if *get_adjusted_prices {
		i++
		go func() {
			results <- GetAndWriteAdjustedPrices(accessToken)
		}()
	}

	if *get_cost_indexes {
		i++
		go func() {
			results <- GetAndWriteCostIndexes(accessToken)
		}()
	}

	if *get_market_orders {
		i++
		go func() {
			results <- GetAndWriteMarketOrders(
				accessToken,
				config.RegionIds,
				config.LocationIds,
			)
		}()
	}

	if *get_assets {
		i++
		go func() {
			results <- GetAndWriteAssets(
				accessToken,
				config.CorporationId,
			)
		}()
	}

	for j := 0; j < i; j++ {
		if err := <-results; err != nil {
			log.Fatal(err)
		}
	}
}
