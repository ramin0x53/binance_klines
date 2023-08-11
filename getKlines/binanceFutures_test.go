package getklines

import (
	"fmt"
	"testing"
	"time"
)

func TestGetKlines(t *testing.T) {
	sample := BinancePerp{Addr: "fapi.binance.com"}
	result, err := sample.getKliens("BTCUSDT", "1h", 1685604839000, 1685691239000, 100)
	if err != nil {
		t.Error(err)
	}
	// fmt.Println(*result)
	for _, i := range *result {
		fmt.Println(i)
	}
	fmt.Println(len(*result))
	if len(*result) != 6 {
		t.Errorf("the length of result must be equal to 6")
	}
	for _, i := range *result {
		fmt.Println(time.Unix(i.OpenTime/1000, 0).UTC())
	}
	fmt.Println((*result)[0])
	fmt.Println((*result)[len(*result)-1])
	fmt.Println(len(*result))
}

// func TestSplitPeriod(t *testing.T) {
// 	a, b, c := split[int](66, 33)
// 	fmt.Println(a)
// 	fmt.Println(b)
// 	fmt.Println(c)
// 	if a != 2 && b != 22 {
// 		t.Errorf("split period not working")
// 	}
// }

func TestKlines(t *testing.T) {
	sample := BinancePerp{Addr: "fapi.binance.com"}
	result, err := sample.Klines("BTCUSDT", "5m", 1685604839000, 1688283239000, 10)
	if err != nil {
		t.Error(err)
	}
	for _, i := range *result {
		fmt.Println(time.Unix(i.OpenTime/1000, 0).UTC())
	}
	fmt.Printf("length ==> %d\n", len(*result))
	// fmt.Println(*result)
}
