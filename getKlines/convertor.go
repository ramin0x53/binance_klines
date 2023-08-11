package getklines

import "fmt"

func ConvertKlinefToJSONArray(klines *[]Klinef) *[][]interface{} {
	var jsonArray [][]interface{}

	for _, k := range *klines {
		jsonArray = append(jsonArray, []interface{}{
			k.OpenTime,
			fmt.Sprintf("%.8f", k.Open),
			fmt.Sprintf("%.8f", k.High),
			fmt.Sprintf("%.8f", k.Low),
			fmt.Sprintf("%.8f", k.Close),
			fmt.Sprintf("%.17f", k.Volume),
			k.CloseTime,
			fmt.Sprintf("%.17f", k.QuoteAssetVolume),
			k.TradeNum,
			fmt.Sprintf("%.17f", k.TakerBuyBaseAssetVolume),
			fmt.Sprintf("%.17f", k.TakerBuyQuoteAssetVolume),
		})
	}

	return &jsonArray
}
