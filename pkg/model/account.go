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

package model

//=============================================================================

type StrategySet struct {
	Strategies map[string]*Strategy
	CurrStrategy *Strategy
}

//=============================================================================

type Strategy struct {
	Name   string
	Ticker string

	DailyInfo []*DailyInfo
}

//=============================================================================

type DailyInfo struct {
	Day         int
	OpenProfit  float64
	Position    int
	GapValue    float64
	TrueRange   float64
	NumTrades   int
}

//=============================================================================

func NewStrategySet() *StrategySet {
	ss := &StrategySet{}
	ss.Strategies = map[string]*Strategy{}
	return ss
}

//=============================================================================

func NewStrategy() *Strategy {
	str := &Strategy{}
	str.DailyInfo = []*DailyInfo{}
	return str
}

//=============================================================================

func NewDailyInfo() *DailyInfo {
	return &DailyInfo{}
}

//=============================================================================
