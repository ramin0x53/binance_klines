package main

import (
	getklines "binance_klines/getKlines"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Request struct {
	Symbol       string `json:"symbol" binding:"required"`
	Interval     string `json:"interval" binding:"required"`
	StartTime    int64  `json:"startTime" binding:"required"`
	EndTime      int64  `json:"endTime" binding:"required"`
	ThreadsCount int    `json:"tCount" binding:"required"`
}

func main() {
	os.Setenv("HTTP_PROXY", "http://localhost:2081")
	os.Setenv("HTTPS_PROXY", "http://localhost:2081")
	// gin.SetMode(gin.ReleaseMode)
	binancePerp := getklines.BinancePerp{Addr: "fapi.binance.com"}
	binanceSpot := getklines.BinanceSpot{Addr: "api.binance.com"}
	server := gin.Default()
	server.POST("/futures", func(c *gin.Context) {
		var input Request
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		klines, err := binancePerp.KlinesArray(input.Symbol, input.Interval, input.StartTime, input.EndTime, input.ThreadsCount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("length ==> ", len(*klines))
		c.JSON(http.StatusOK, klines)
	})
	server.POST("/spot", func(c *gin.Context) {
		var input Request
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		klines, err := binanceSpot.KlinesArray(input.Symbol, input.Interval, input.StartTime, input.EndTime, input.ThreadsCount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("length ==> ", len(*klines))
		c.JSON(http.StatusOK, klines)
	})
	server.Run()
}
