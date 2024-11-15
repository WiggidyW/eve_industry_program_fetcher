package main

import (
	"flag"
	"log"
)

func main() {
	get_adjusted_prices := flag.Bool("adjusted_prices", false, "Get adjusted prices")
	get_cost_indices := flag.Bool("cost_indices", false, "Get cost indices")
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

	log.Println("Authenticated")

	i := 0
	results := make(chan error, 4)

	if *get_adjusted_prices {
		i++
		go func() {
			results <- GetAndWriteAdjustedPrices(accessToken)
			log.Println("Wrote adjusted prices")
		}()
	}

	if *get_cost_indices {
		i++
		go func() {
			results <- GetAndWriteCostIndices(accessToken)
			log.Println("Wrote cost indices")
		}()
	}

	if *get_market_orders {
		i++
		go func() {
			results <- GetAndWriteMarketOrders(
				accessToken,
				config.LocationIds,
				config.RegionIds,
			)
			log.Println("Wrote market orders")
		}()
	}

	if *get_assets {
		i++
		go func() {
			results <- GetAndWriteAssets(
				accessToken,
				config.CorporationId,
			)
			log.Println("Wrote assets")
		}()
	}

	for j := 0; j < i; j++ {
		if err := <-results; err != nil {
			log.Fatal(err)
		}
	}
}
