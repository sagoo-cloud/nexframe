package convert

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/errors/gcode"
	"github.com/sagoo-cloud/nexframe/utils/errors/gerror"
	"regexp"
	"strconv"
	"time"
)

// Time converts `any` to time.Time.
func Time(any interface{}, format ...string) time.Time {
	// It's already this type.
	if len(format) == 0 {
		if v, ok := any.(time.Time); ok {
			return v
		}
	}
	return time.Time{}
}

func Duration(any interface{}) time.Duration {
	// It's already this type.
	if v, ok := any.(time.Duration); ok {
		return v
	}
	s := String(any)
	if !utils.IsNumeric(s) {
		d, _ := ParseDuration(s)
		return d
	}
	return time.Duration(Int64(any))
}
func ParseDuration(s string) (duration time.Duration, err error) {
	var (
		num int64
	)
	if utils.IsNumeric(s) {
		num, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			err = gerror.WrapCodef(gcode.CodeInvalidParameter, err, `strconv.ParseInt failed for string "%s"`, s)
			return 0, err
		}
		return time.Duration(num), nil
	}

	// 编译正则表达式
	re, err := regexp.Compile(`^([\-\d]+)[dD](.*)$`)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return
	}

	// 使用FindStringSubmatch进行匹配
	match := re.FindStringSubmatch(s)
	if match == nil {
		fmt.Println("No match found")
		return
	}

	if len(match) == 3 {
		num, err = strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			err = gerror.WrapCodef(gcode.CodeInvalidParameter, err, `strconv.ParseInt failed for string "%s"`, match[1])
			return 0, err
		}
		s = fmt.Sprintf(`%dh%s`, num*24, match[2])
		duration, err = time.ParseDuration(s)
		if err != nil {
			err = gerror.WrapCodef(gcode.CodeInvalidParameter, err, `time.ParseDuration failed for string "%s"`, s)
		}
		return
	}
	duration, err = time.ParseDuration(s)
	err = gerror.WrapCodef(gcode.CodeInvalidParameter, err, `time.ParseDuration failed for string "%s"`, s)
	return
}
