// Package gcharset 实现字符集转换功能。
//
// 支持的字符集:
//
// 中文: GBK/GB18030/GB2312/Big5
// 日语: EUCJP/ISO2022JP/ShiftJIS
// 韩语: EUCKR
// Unicode: UTF-8/UTF-16/UTF-16BE/UTF-16LE
// 其他: macintosh/IBM*/Windows*/ISO-*
package gcharset

import (
	"bytes"
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/errors/gcode"
	gerror2 "github.com/sagoo-cloud/nexframe/utils/errors/gerror"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"io"
	"strings"
	"sync"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

var (
	// charsetAliases 用于字符集别名映射
	charsetAliases = map[string]string{
		"HZGB2312": "HZ-GB-2312",
		"hzgb2312": "HZ-GB-2312",
		"GB2312":   "GBK",
		"gb2312":   "GBK",
	}

	// encodingMap 用于直接映射字符集到编码对象
	encodingMap = map[string]encoding.Encoding{
		"UTF-8":    unicode.UTF8,
		"UTF-16":   unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
		"UTF-16LE": unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
		"UTF-16BE": unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM),
		"GBK":      simplifiedchinese.GBK,
		"GB18030":  simplifiedchinese.GB18030,
		"Big5":     traditionalchinese.Big5,
		"EUCJP":    japanese.EUCJP,
		"ShiftJIS": japanese.ShiftJIS,
		"EUCKR":    korean.EUCKR,
	}

	// encodingCache 用于缓存编码对象，提高性能
	encodingCache sync.Map
)

// ErrUnsupportedCharset 表示不支持的字符集错误
var ErrUnsupportedCharset = errors.New("unsupported charset")

// Supported 返回是否支持指定的字符集
func Supported(charset string) bool {
	charset = strings.ToUpper(charset)
	if _, ok := encodingMap[charset]; ok {
		return true
	}
	return getEncoding(charset) != nil
}

// Convert 将src从srcCharset编码转换为dstCharset编码，并返回转换后的字符串
// 如果转换失败，则返回src作为dst
func Convert(dstCharset, srcCharset string, src string) (dst string, err error) {
	if dstCharset == srcCharset {
		return src, nil
	}

	// 将src转换为UTF-8
	if srcCharset != "UTF-8" {
		src, err = ToUTF8(srcCharset, src)
		if err != nil {
			return "", err
		}
	}

	// 将UTF-8转换为目标编码
	if dstCharset != "UTF-8" {
		return UTF8To(dstCharset, src)
	}

	return src, nil
}

// ToUTF8 将src从srcCharset编码转换为UTF-8，并返回转换后的字符串
func ToUTF8(srcCharset string, src string) (dst string, err error) {
	e := getEncoding(srcCharset)
	if e == nil {
		return "", gerror2.Wrapf(ErrUnsupportedCharset, "unsupported srcCharset %q", srcCharset)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, transform.NewReader(bytes.NewReader([]byte(src)), e.NewDecoder()))
	if err != nil {
		return "", gerror2.Wrapf(err, "convert string from %q to UTF-8 failed", srcCharset)
	}
	return buf.String(), nil
}

// UTF8To 将src从UTF-8编码转换为dstCharset，并返回转换后的字符串
func UTF8To(dstCharset string, src string) (dst string, err error) {
	e := getEncoding(dstCharset)
	if e == nil {
		return "", gerror2.Wrapf(ErrUnsupportedCharset, "unsupported dstCharset %q", dstCharset)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, transform.NewReader(bytes.NewReader([]byte(src)), e.NewEncoder()))
	if err != nil {
		return "", gerror2.Wrapf(err, "convert string from UTF-8 to %q failed", dstCharset)
	}
	return buf.String(), nil
}

// getEncoding 返回指定字符集的encoding.Encoding接口对象
// 如果字符集不支持，则返回nil
func getEncoding(charset string) encoding.Encoding {
	charset = strings.ToUpper(charset)

	// 检查并使用字符集别名
	if alias, ok := charsetAliases[charset]; ok {
		charset = alias
	}

	// 首先检查直接映射
	if enc, ok := encodingMap[charset]; ok {
		return enc
	}

	// 尝试从缓存中获取编码对象
	if enc, ok := encodingCache.Load(charset); ok {
		return enc.(encoding.Encoding)
	}

	// 如果缓存中不存在，则尝试创建新的编码对象
	enc, err := getEncodingFromIANA(charset)
	if err != nil {
		return nil
	}

	// 将新创建的编码对象存入缓存
	encodingCache.Store(charset, enc)
	return enc
}

func getEncodingFromIANA(charset string) (encoding.Encoding, error) {
	return ianaindex.MIME.Encoding(charset)
}

// MaxInputSize 定义了允许处理的最大输入大小（例如：10MB）
const MaxInputSize = 10 * 1024 * 1024 // 10MB

// ConvertWithSizeLimit 在进行转换之前检查输入大小，防止处理过大的输入
func ConvertWithSizeLimit(dstCharset, srcCharset string, src string) (dst string, err error) {
	if len(src) > MaxInputSize {
		return "", gerror2.NewCodef(gcode.CodeInvalidParameter, "input size exceeds limit of %d bytes", MaxInputSize)
	}
	return Convert(dstCharset, srcCharset, src)
}
