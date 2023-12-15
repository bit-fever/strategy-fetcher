//=============================================================================
/*
Copyright © 2023 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package core

import (
	"bufio"
	"github.com/bit-fever/strategy-fetcher/pkg/model"
	"github.com/bit-fever/strategy-fetcher/pkg/model/config"
	"golang.org/x/exp/maps"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

//=============================================================================

var strategies map[string]*model.Strategy

//=============================================================================

func GetStrategies() []*model.Strategy {
	return maps.Values(strategies)
}

//=============================================================================

func StartPeriodicScan(cfg *config.Config) *time.Ticker {

	ticker := time.NewTicker(cfg.Scan.PeriodHour * time.Hour)

	go func() {
		time.Sleep(2 * time.Second)
		run(cfg)

		for range ticker.C {
			run(cfg)
		}
	}()

	return ticker
}

//=============================================================================

func run(cfg *config.Config) {
	dir := cfg.Scan.Dir
	log.Println("Fetching files from: " + dir)

	files, error := os.ReadDir(dir)

	if error != nil {
		log.Println("Scan error: ", error)
	} else {
		ss := model.NewStrategySet()
		for _, entry := range files {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".data") {
				handleFile(ss, dir, entry.Name())
			}
		}
		strategies = ss.Strategies
	}
}

//=============================================================================

func handleFile(ss *model.StrategySet, dir string, fileName string) {
	log.Println("Handling: " + fileName)

	path := dir + string(os.PathSeparator) + fileName

	file, err := os.Open(path)

	if err != nil {
		log.Println("Cannot open file for reading: " + path + " (cause is: " + err.Error() + " )")
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		handleLine(ss, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println("Cannot scan file: " + path + " (cause is: " + err.Error() + " )")
	}
}

//=============================================================================

func handleLine(ss *model.StrategySet, line string) {
	tokens := strings.Split(line, "|")

	switch tokens[0] {
	case INFO:
		handleInfo(ss, tokens)
	case DAILY:
		handleDaily(ss, tokens)
	default:
		log.Println("Skipping unknown token: " + tokens[0])
	}
}

//=============================================================================

func handleInfo(ss *model.StrategySet, tokens []string) {
	ticker   := tokens[1]
	strategy := tokens[2]

	str, ok := ss.Strategies[strategy]

	if !ok {
		str = model.NewStrategy()
		str.Name   = strategy
		str.Ticker = ticker
		ss.Strategies[strategy] = str
		ss.CurrStrategy = str
	}
}

//=============================================================================

func handleDaily(ss *model.StrategySet, tokens []string) {
	strDay          := tokens[1]
	strOpenEquity   := tokens[2]
	strClosedEquity := tokens[3]
	strPosition     := tokens[4]
	strNumTrades    := tokens[5]

	day := convertDate(strDay)

	if day == 0 {
		return
	}

	di := model.NewDailyInfo()
	ss.CurrStrategy.DailyInfo = append(ss.CurrStrategy.DailyInfo, di)

	di.Day = day

	convertOpenProfit  (di, strOpenEquity)
	convertClosedProfit(di, strClosedEquity)
	convertPosition    (di, strPosition)
	convertNumTrades   (di, strNumTrades)
}

//=============================================================================

func convertDate(strDate string) int {
	tokens := strings.Split(strDate, "/")

	if len(tokens) != 3 {
		log.Println("Bad format for day: " + strDate)
		return 0
	}

	value, err := strconv.ParseInt(tokens[2]+tokens[1]+tokens[0], 10, 32)

	if err != nil {
		log.Println("Internal error: cannot parse day as int --> " + strDate)
		return 0
	}

	if value < 20000000 || value > 30000000 {
		log.Println("Bad value for day: " + strDate)
		return 0
	}

	return int(value)
}

//=============================================================================

func convertOpenProfit(di *model.DailyInfo, strValue string) {
	value, err := strconv.ParseFloat(strValue, 64)

	if err != nil {
		log.Println("Cannot convert open profit: " + strValue)
	} else {
		di.OpenProfit = value
	}
}

//=============================================================================

func convertClosedProfit(di *model.DailyInfo, strValue string) {
	value, err := strconv.ParseFloat(strValue, 64)

	if err != nil {
		log.Println("Cannot convert closed profit: " + strValue)
	} else {
		di.ClosedProfit = value
	}
}

//=============================================================================

func convertPosition(di *model.DailyInfo, strValue string) {
	value, err := strconv.ParseInt(strValue, 10, 32)

	if err != nil {
		log.Println("Cannot convert position: " + strValue)
	} else {
		di.Position = int(value)
	}
}

//=============================================================================

func convertNumTrades(di *model.DailyInfo, strValue string) {
	value, err := strconv.ParseInt(strValue, 10, 32)

	if err != nil {
		log.Println("Cannot convert num trades: " + strValue)
	} else {
		di.NumTrades = int(value)
	}
}

//=============================================================================
