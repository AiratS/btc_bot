package main

type Buffer struct {
	candles []Candle
	maxSize int
}

func NewBuffer(maxSize int) Buffer {
	return Buffer{maxSize: maxSize}
}

func (buffer *Buffer) AddCandle(candle Candle) {
	buffer.candles = append(buffer.candles, candle)
	realSize := len(buffer.candles)
	if realSize <= buffer.maxSize {
		return
	}

	tempCandles := buffer.candles[1:]
	buffer.candles = tempCandles
}

func (buffer *Buffer) GetCandles() []Candle {
	return buffer.candles
}

func (buffer *Buffer) GetBackCandle(nBack int) Candle {
	lastIdx := len(buffer.candles) - 1

	return buffer.candles[lastIdx-nBack]
}

func (buffer *Buffer) GetLastCandle() Candle {
	//return Candle{
	//	Symbol:     CANDLE_SYMBOL,
	//	CloseTime:  time.Now().Format("2006-01-02 15:04:05"),
	//	ClosePrice: 23398.11,
	//}
	return buffer.candles[len(buffer.candles)-1]
}

func (buffer *Buffer) GetLastCandleClosePrice() float64 {
	candle := buffer.GetLastCandle()
	return candle.GetPrice()
}
