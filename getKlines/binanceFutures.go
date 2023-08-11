package getklines

import (
	"encoding/json"
	"errors"
	"fmt"

	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TimeFrames map[string]int64

func (t TimeFrames) Sec(tf string) int64 {
	return t[tf]
}

func (t TimeFrames) MilSec(tf string) int64 {
	return t[tf] * 1000
}

var tframes = TimeFrames{
	"1m":  60,
	"3m":  180,
	"5m":  300,
	"15m": 900,
	"30m": 1800,
	"1h":  3600,
	"2h":  7200,
	"4h":  14400,
	"6h":  21600,
	"8h":  28800,
	"12h": 43200,
	"1d":  86400,
	"3d":  259200,
	"1w":  604800,
}

var maxLimit int64 = 1500

func split[V int32 | int64 | int16 | int8 | int](length, limit V) (V, V, [][]V) {
	splitRanges := [][]V{}
	divideResult := (length / limit)
	divideRemainResult := (length % limit)
	if length >= limit {
		for i := V(0); i < divideResult*limit; i += limit {
			splitRanges = append(splitRanges, []V{i, i + limit})
		}
	}
	if divideRemainResult > 0 {
		splitRanges = append(splitRanges, []V{divideResult * limit, (divideResult * limit) + divideRemainResult})
	}
	return divideResult, divideRemainResult, splitRanges
}

type Klinef struct {
	OpenTime                 int64
	Open                     float64
	High                     float64
	Low                      float64
	Close                    float64
	Volume                   float64
	CloseTime                int64
	QuoteAssetVolume         float64
	TradeNum                 int64
	TakerBuyBaseAssetVolume  float64
	TakerBuyQuoteAssetVolume float64
}

type klinesRequest struct {
	symbol    string
	interval  string
	startTime int64
	endTime   int64
	limit     int64
}

type klinesResponse struct {
	result *[]Klinef
	err    error
}

type BinancePerp struct{ Addr string }

func NewBinancePerp(addr string) *BinancePerp {
	return &BinancePerp{Addr: addr}
}

func (b *BinancePerp) getKliens(symbol string, interval string, startTime int64, endTime int64, limit int64) (*[]Klinef, error) {
	url := url.URL{Scheme: "https", Host: b.Addr, Path: "/fapi/v1/klines", RawQuery: fmt.Sprintf("symbol=%s&interval=%s&startTime=%d&endTime=%d&limit=%d", strings.ToUpper(symbol), interval, startTime, endTime, limit)}
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

func (b *BinancePerp) getKlinesWorker(klinesReq <-chan klinesRequest, klinesRes chan<- klinesResponse) {
	for j := range klinesReq {
		klines, err := b.getKliens(j.symbol, j.interval, j.startTime, j.endTime, j.limit)
		result := klinesResponse{
			result: klines,
			err:    err,
		}
		klinesRes <- result
	}
}

func (b *BinancePerp) Klines(symbol, interval string, startTime, endTime int64, threadsCount int) (*[]Klinef, error) {
	if !isTimeframeExist(interval, tframes) {
		return nil, errors.New("this interval doesn't exist")
	}
	endTime += tframes[interval]
	timeDifference := endTime - startTime
	limitInTime := tframes.MilSec(interval) * maxLimit
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
			limit:     maxLimit,
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

func (b *BinancePerp) KlinesArray(symbol, interval string, startTime, endTime int64, threadsCount int) (*[][]interface{}, error) {
	klines, err := b.Klines(symbol, interval, startTime, endTime, threadsCount)
	if err != nil {
		return nil, err
	}
	return ConvertKlinefToJSONArray(klines), nil
}

func sortWorkersResult(data []klinesResponse) *[]Klinef {
	indexHolder := make(map[int64]int)
	arrayOpenTimes := []int64{}
	for i, a := range data {
		indexHolder[(*a.result)[0].OpenTime] = i
		arrayOpenTimes = append(arrayOpenTimes, (*a.result)[0].OpenTime)
	}

	arrayOpenTimes = quicksort(arrayOpenTimes)
	allKlines := []Klinef{}
	for _, n := range arrayOpenTimes {
		allKlines = append(allKlines, *data[indexHolder[n]].result...)
	}
	return &allKlines
}

func quicksort(a []int64) []int64 {
	if len(a) < 2 {
		return a
	}

	left, right := 0, len(a)-1

	pivot := rand.Int() % len(a)

	a[pivot], a[right] = a[right], a[pivot]

	for i := range a {
		if a[i] < a[right] {
			a[left], a[i] = a[i], a[left]
			left++
		}
	}

	a[left], a[right] = a[right], a[left]

	quicksort(a[:left])
	quicksort(a[left+1:])

	return a
}

func StringToFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func isTimeframeExist(key string, m TimeFrames) bool {
	_, ok := m[key]
	return ok
}
