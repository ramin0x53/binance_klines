package getklines

import (
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
	"net/url"
	"strings"
)

var maxLimitSpot int64 = 1000

type BinanceSpot struct{ Addr string }

func NewBinanceSpot(addr string) *BinanceSpot {
	return &BinanceSpot{Addr: addr}
}

func (b *BinanceSpot) getKliens(symbol string, interval string, startTime int64, endTime int64, limit int64) (*[]Klinef, error) {
	url := url.URL{Scheme: "https", Host: b.Addr, Path: "/api/v3/klines", RawQuery: fmt.Sprintf("symbol=%s&interval=%s&startTime=%d&endTime=%d&limit=%d", strings.ToUpper(symbol), interval, startTime, endTime, limit)}
	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body := [][]interface{}{}
	json.NewDecoder(resp.Body).Decode(&body)
	var klinesf []Klinef
	for _, i := range body {
		klinesf = append(klinesf, Klinef{
			OpenTime:                 int64(i[0].(float64)),
			Open:                     StringToFloat64(i[1].(string)),
			High:                     StringToFloat64(i[2].(string)),
			Low:                      StringToFloat64(i[3].(string)),
			Close:                    StringToFloat64(i[4].(string)),
			Volume:                   StringToFloat64(i[5].(string)),
			CloseTime:                int64(i[6].(float64)),
			QuoteAssetVolume:         StringToFloat64(i[7].(string)),
			TradeNum:                 int64(i[8].(float64)),
			TakerBuyBaseAssetVolume:  StringToFloat64(i[9].(string)),
			TakerBuyQuoteAssetVolume: StringToFloat64(i[10].(string)),
		})
	}

	return &klinesf, nil
}

func (b *BinanceSpot) getKlinesWorker(klinesReq <-chan klinesRequest, klinesRes chan<- klinesResponse) {
	for j := range klinesReq {
		klines, err := b.getKliens(j.symbol, j.interval, j.startTime, j.endTime, j.limit)
		result := klinesResponse{
			result: klines,
			err:    err,
		}
		klinesRes <- result
	}
}

func (b *BinanceSpot) Klines(symbol, interval string, startTime, endTime int64, threadsCount int) (*[]Klinef, error) {
	if !isTimeframeExist(interval, tframes) {
		return nil, errors.New("this interval doesn't exist")
	}
	endTime += tframes[interval]
	timeDifference := endTime - startTime
	limitInTime := tframes.MilSec(interval) * maxLimitSpot
	_, _, ranges := split[int64](timeDifference, limitInTime)

	for i := range ranges {
		ranges[i] = []int64{ranges[i][0] + startTime, ranges[i][1] + startTime}
	}

	rangeCount := len(ranges)
	jobs := make(chan klinesRequest, rangeCount)
	results := make(chan klinesResponse, rangeCount)

	for w := 0; w < threadsCount; w++ {
		go b.getKlinesWorker(jobs, results)
	}

	for j := 0; j < rangeCount; j++ {
		job := klinesRequest{
			symbol:    symbol,
			interval:  interval,
			startTime: ranges[j][0],
			endTime:   ranges[j][1],
			limit:     maxLimitSpot,
		}

		jobs <- job
	}
	close(jobs)

	workerResult := []klinesResponse{}
	for a := 0; a < rangeCount; a++ {
		res := <-results
		if res.err != nil {
			return nil, res.err
		}
		workerResult = append(workerResult, res)
	}
	close(results)

	return sortWorkersResult(workerResult), nil
}

func (b *BinanceSpot) KlinesArray(symbol, interval string, startTime, endTime int64, threadsCount int) (*[][]interface{}, error) {
	klines, err := b.Klines(symbol, interval, startTime, endTime, threadsCount)
	if err != nil {
		return nil, err
	}
	return ConvertKlinefToJSONArray(klines), nil
}
