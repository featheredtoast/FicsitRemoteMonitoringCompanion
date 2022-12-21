package exporter

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"strconv"
	"time"
)

type DroneStationCollector struct {
	FRMAddress string
}

type DroneStationDetails struct {
	Id int `json:"ID"`
	Location Location `json:"location"`
	HomeStation string `json:HomeStation`
	PariedStation string `json:"PairedStation"`
	DroneStatus string `json:"DroneStatus"`
	AvgIncRate float64 `json:"AvgIncRate"`
	AvgIncStack float64 `json:"AvgIncStack"`
	AvgOutRate float64 `json:"AvgOutRate"`
	AvgOutStack float64 `json"AvgOutStack"`
	AvgRndTrip string `json:"AvgRndTrip"`
	AvgTotalIncRate float64 `json:"AvgTotalIncRate"`
	AvgTotalIncStack float64 `json:"AvgTotalIncStack"`
	AvgTotalOutRate float64 `json:"AvgTotalOutRate"`
	AvgTotalOutStack float64 `json:"AvgTotalOutStack"`
	AvgTripIncAmt float64 `json:"AvgTripIncAmt"`
	EstRndTrip string `json:"EstRndTrip"`
	EstTotalTransRate float64 `json:"EstTotalTransRate"`
	EstTransRate float64 `json:"EstTransRate"`
	EstLatestTotalIncStack float64 `json:"EstLatestTotalIncStack"`
	EstLatestTotalOutStack float64 `json:"EstLatestTotalOutStack"`
	LatestIncStack float64 `json:"LatestIncStack"`
	LatestOutStack float64 `json:"LatestOutStack"`
	LatestRndTrip string `json:"LatestRndTrip"`
	LatestTripIncAmt int `json:"LatestTripIncAmt"`
	LatestTripOutAmt int `json:"LatestTripOutAmt"`
	MedianRndTrip string `json:"MedianRndTrip"`
	MedianTripIncAmt float64 `json:"MedianTripIncAmt"`
	MedianTripOutAmt float64 `json:"MedianTripOutAmt"`
	estBatteryRate float64 `json:"EstBatteryRate"`
}
