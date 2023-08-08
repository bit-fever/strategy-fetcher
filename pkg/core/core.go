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

var accounts map[string]*model.Account

//=============================================================================

func GetAccounts() []*model.Account {
	return maps.Values(accounts)
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
		as := model.NewAccountSet()
		for _, entry := range files {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
				handleFile(as, dir, entry.Name())
			}
		}
		accounts = as.Accounts
	}
}

//=============================================================================

func handleFile(as *model.AccountSet, dir string, fileName string) {
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
		handleLine(as, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println("Cannot scan file: " + path + " (cause is: " + err.Error() + " )")
	}
}

//=============================================================================

func handleLine(as *model.AccountSet, line string) {

	tokens := strings.Split(line, "|")

	switch tokens[0] {
	case INFO:
		handleInfo(as, tokens)
	case DAILY:
		handleDaily(as, tokens)
	case LONG_ENTRY,LONG_EXIT,SHORT_ENTRY,SHORT_EXIT,LONG_SHORT,SHORT_LONG:
		handleTrade(as, tokens)
	default:
		log.Println("Skipping unknown token: " + tokens[0])
	}
}

//=============================================================================

func handleInfo(as *model.AccountSet, tokens []string) {
	accName := tokens[1]
	ticker := tokens[2]
	strategy := tokens[3]

	acc, ok := as.Accounts[accName]

	if !ok {
		acc = model.NewAccount()
		acc.Code = accName
		as.Accounts[accName] = acc
		as.CurrAccount = acc
	}

	str, ok := acc.Strategies[strategy]

	if !ok {
		str = model.NewStrategy()
		str.Name = strategy
		str.Ticker = ticker
		acc.Strategies[strategy] = str
		as.CurrStrategy = str
	}
}

//=============================================================================

func handleDaily(as *model.AccountSet, tokens []string) {
	strDay        := tokens[1]
	strOpenEquity := tokens[2]
	strNetProfit  := tokens[3]
	strTrueRange  := tokens[4]
	strNumTrades  := tokens[5]
	strEquity     := tokens[6]
	strBalance    := tokens[7]

	day := convertDate(strDay)

	if day == 0 {
		return
	}

	di := model.NewDailyInfo()
	as.CurrStrategy.DailyInfo = append(as.CurrStrategy.DailyInfo, di)

	di.Day = day

	convertOpenProfit (di, strOpenEquity)
	convertCloseProfit(di, strNetProfit)
	convertTrueRange  (di, strTrueRange)
	convertNumTrades  (di, strNumTrades)
	convertEquity     (di, strEquity)
	convertBalance    (di, strBalance)
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

func convertCloseProfit(di *model.DailyInfo, strValue string) {
	value, err := strconv.ParseFloat(strValue, 64)

	if err != nil {
		log.Println("Cannot convert close profit: " + strValue)
	} else {
		di.CloseProfit = value
	}
}

//=============================================================================

func convertTrueRange(di *model.DailyInfo, strValue string) {
	value, err := strconv.ParseFloat(strValue, 64)

	if err != nil {
		log.Println("Cannot convert true range: " + strValue)
	} else {
		di.TrueRange = value
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

func convertBalance(di *model.DailyInfo, strValue string) {
	value, err := strconv.ParseFloat(strValue, 64)

	if err != nil {
		log.Println("Cannot convert balance: " + strValue)
	} else {
		di.Balance = value
	}
}

//=============================================================================

func convertEquity(di *model.DailyInfo, strValue string) {
	value, err := strconv.ParseFloat(strValue, 64)

	if err != nil {
		log.Println("Cannot convert equity: " + strValue)
	} else {
		di.Equity = value
	}
}

//=============================================================================

func handleTrade(as *model.AccountSet, tokens []string) {
	strType       := tokens[0]
	strDay        := tokens[1]
	strTime       := tokens[2]
	strPosition   := tokens[3]
	strPrice      := tokens[4]
	strPosAtBrk   := tokens[5]
	strPriceAtBrk := tokens[6]

	day := convertDate(strDay)

	if day == 0 {
		return
	}

	ti := model.NewTradeInfo()
	as.CurrStrategy.TradeInfo = append(as.CurrStrategy.TradeInfo, ti)

	ti.Type= strType
	ti.Day = day

	convertTime           (ti, strTime)
	convertPosition        (ti, strPosition)
	convertPrice           (ti, strPrice)
	convertPositionAtBroker(ti, strPosAtBrk)
	convertPriceAtBroker   (ti, strPriceAtBrk)
}

//=============================================================================

func convertTime(ti *model.TradeInfo, strValue string) {
	value, err := strconv.ParseInt(strValue, 10, 32)

	if err != nil {
		log.Println("Cannot convert time: " + strValue)
	} else {
		ti.Time = int(value)
	}
}

//=============================================================================

func convertPosition(ti *model.TradeInfo, strValue string) {
	value, err := strconv.ParseInt(strValue, 10, 32)

	if err != nil {
		log.Println("Cannot convert position: " + strValue)
	} else {
		ti.Position = int(value)
	}
}

//=============================================================================

func convertPrice(ti *model.TradeInfo, strValue string) {
	value, err := strconv.ParseFloat(strValue, 64)

	if err != nil {
		log.Println("Cannot convert price: " + strValue)
	} else {
		ti.Price = value
	}
}

//=============================================================================

func convertPositionAtBroker(ti *model.TradeInfo, strValue string) {
	value, err := strconv.ParseInt(strValue, 10, 32)

	if err != nil {
		log.Println("Cannot convert position at broker: " + strValue)
	} else {
		ti.PositionAtBroker = int(value)
	}
}

//=============================================================================

func convertPriceAtBroker(ti *model.TradeInfo, strValue string) {
	value, err := strconv.ParseFloat(strValue, 64)

	if err != nil {
		log.Println("Cannot convert price at broker: " + strValue)
	} else {
		ti.PriceAtBroker = value
	}
}

//=============================================================================
