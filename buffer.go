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
