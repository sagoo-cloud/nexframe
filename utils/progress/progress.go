package progress

import (
	"fmt"
)

// Bar 表示一个进度条
type Bar struct {
	percent int    // 当前进度百分比
	cur     int    // 当前进度位置
	total   int    // 总数
	rate    string // 进度条
	graph   string // 显示符号
}

// NewProgressBar 创建一个新的进度条
func NewProgressBar(start, total int) *Bar {
	bar := &Bar{
		cur:   start,
		total: total,
		graph: "█",
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph
	}
	return bar
}

// getPercent 获取百分比
func (bar *Bar) getPercent() int {
	if bar.cur >= bar.total {
		return 100
	}
	return int(float32(bar.cur) * 100 / float32(bar.total))
}

// Play 显示进度
func (bar *Bar) Play(cur int) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
	fmt.Printf("\r[%-50s]%3d%%  %8d/%d", bar.rate, bar.percent, bar.cur, bar.total)
}

// Finish 完成进度
func (bar *Bar) Finish() {
	if bar.cur < bar.total {
		bar.Play(bar.total) // 确保显示100%
	}
	fmt.Println()
}
