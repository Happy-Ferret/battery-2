package battery

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/fatih/color"
)

// LENGTH: 8
var chars = []rune{
	'▏',
	'▎',
	'▍',
	'▌',
	'▋',
	'▊',
	'▉',
	'█',
}

type Bar struct {
	buffer      bytes.Buffer
	gauge       []rune
	gaugeWidth  int
	width       int
	nowVal      int
	totalVal    int
	charLen     int
	format      string
	prefix      rune
	postfix     rune
	charge      string
	ShowPercent bool
	ShowCounter bool
	Showthunder bool
	EnableColor bool
	EnableTmux  bool
}

func digit(num int) string {
	return strconv.Itoa(int(math.Log10(float64(num))) + 1)
}

func New(total int) *Bar {
	if total <= 0 {
		panic(errors.New("Please specify total size that is greater than zero"))
	}
	bar := &Bar{
		buffer:      bytes.Buffer{},
		totalVal:    total,
		nowVal:      -1,
		charLen:     len(chars),
		format:      "%s",
		prefix:      '|',
		postfix:     '|',
		charge:      "⚡︎",
		ShowPercent: true,
		ShowCounter: true,
		Showthunder: false,
		EnableColor: false,
		EnableTmux:  true,
	}
	return bar.SetWidth(3)
}

func (bar *Bar) SetPrefix(char rune) *Bar {
	bar.prefix = char
	return bar
}

func (bar *Bar) SetPostfix(char rune) *Bar {
	bar.postfix = char
	return bar
}

func (bar *Bar) SetWidth(width int) *Bar {
	bar.width = width
	// +1 for postfix
	bar.gaugeWidth = width + 1
	// +1 for prefix
	bar.gauge = make([]rune, bar.gaugeWidth+1, bar.gaugeWidth+1)
	return bar
}

func (bar *Bar) Set(set int) *Bar {
	bar.nowVal = set
	return bar
}

func (bar *Bar) Run() {
	bar.writer()
}

func (bar *Bar) writer() {
	if bar.ShowPercent {
		bar.format = "%3d%%" + bar.format
	}

	if bar.ShowCounter {
		digit := digit(bar.totalVal)
		bar.format += " %" + digit + "d/%" + digit + "d"
	}

	// for thunder
	bar.format = "%s" + bar.format

	if bar.nowVal <= bar.totalVal {
		bar.print()
	}

}

func (bar *Bar) print() {
	frac := float64(bar.nowVal) / float64(bar.totalVal)
	barLen, fracBarLen := bar.divmod(frac)

	// append prefix
	bar.gauge[0] = bar.prefix

	// append full block
	for i := 1; i < barLen; i++ {
		bar.gauge[i] = chars[bar.charLen-1]
	}

	// append lower block
	bar.gauge[barLen] = chars[fracBarLen]

	// padding with whitespace
	for i := barLen + 1; i < bar.gaugeWidth; i++ {
		bar.gauge[i] = ' '
	}

	// append postfix
	bar.gauge[bar.gaugeWidth] = bar.postfix

	bar.write(frac)
}

func (bar *Bar) write(frac float64) {
	var args []interface{}
	percent := int(frac * 100)

	if bar.Showthunder {
		args = append(args, bar.charge)
	} else {
		args = append(args, "  ")
	}

	if bar.ShowPercent {
		args = append(args, percent)
	}

	args = append(args, string(bar.gauge))

	if bar.ShowCounter {
		args = append(args, bar.nowVal)
		args = append(args, bar.totalVal)
	}

	if bar.EnableColor {
		if bar.EnableTmux {
			bar.colorTmuxPrint(percent, args...)
		} else {
			bar.colorPrint(percent, args...)
		}
	} else {
		fmt.Fprintf(&bar.buffer, bar.format, args...)
	}
	color.Output.Write(bar.buffer.Bytes())
}

func (bar *Bar) colorTmuxPrint(percent int, args ...interface{}) {
	if percent >= 60 {
		bar.format = "#[fg=green]" + bar.format
	} else if 20 <= percent && percent < 60 {
		bar.format = "#[fg=yellow]" + bar.format
	} else {
		bar.format = "#[fg=red]" + bar.format
	}
	fmt.Fprintf(&bar.buffer, bar.format+"#[default]", args...)
}

func (bar *Bar) colorPrint(percent int, args ...interface{}) {
	if percent >= 60 {
		bar.buffer.WriteString(color.GreenString(bar.format, args...))
	} else if 20 <= percent && percent < 60 {
		bar.buffer.WriteString(color.YellowString(bar.format, args...))
	} else {
		bar.buffer.WriteString(color.RedString(bar.format, args...))
	}
}

func (bar *Bar) divmod(frac float64) (int, int) {
	// Over 100%
	if frac >= 1.0 {
		return bar.width, bar.charLen - 1
	}
	pre := int(frac * float64(bar.width) * float64(bar.charLen))
	return pre/bar.charLen + 1, pre % bar.charLen
}
