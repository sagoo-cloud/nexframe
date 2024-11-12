package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sagoo-cloud/nexframe/i18n"
	"github.com/tealeg/xlsx/v3"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ToMultiSheetExcel 生成io.ReadSeeker  参数 titleList 为Excel表头，dataList 为数据，data 键为sheet
func ToMultiSheetExcel(data map[string][]any) (content io.ReadSeeker) {
	// 生成一个新的文件
	file := xlsx.NewFile()

	for sheetName, dataList := range data {
		// 添加sheet页
		sheet, err := file.AddSheet(sheetName)
		if err != nil {
			log.Printf("Failed to add sheet: %v", err)
			continue
		}

		// 插入表头
		titleRow := sheet.AddRow()

		//获取表头
		objType := reflect.TypeOf(dataList[0])
		elem := objType.Elem()
		var titleList []string
		if elem.Kind() == reflect.Struct {
			for i := 1; i <= elem.NumField(); i++ {
				field := elem.Field(i - 1)
				if field.Name != "PageReq" {
					if field.Tag != "" && field.Tag.Get("dc") != "" {
						titleList = append(titleList, i18n.T(field.Tag.Get("dc")))
					} else {
						titleList = append(titleList, i18n.T(field.Name))
					}
				}
			}
		}

		// 创建表头样式
		headerStyle := xlsx.NewStyle()
		headerStyle.Font = *xlsx.NewFont(10, "Arial")
		headerStyle.Font.Color = "000000"
		headerStyle.Fill.PatternType = "solid"
		headerStyle.Fill.FgColor = "cfe2f3"
		headerStyle.Alignment.Horizontal = "center"
		headerStyle.Alignment.Vertical = "center"

		for _, v := range titleList {
			cell := titleRow.AddCell()
			cell.Value = v
			cell.SetStyle(headerStyle)
		}

		// 插入内容
		for _, v := range dataList {
			row := sheet.AddRow()
			writeStructToRow(row, v)
		}
	}

	var buffer bytes.Buffer
	err := file.Write(&buffer)
	if err != nil {
		log.Printf("Failed to write file: %v", err)
		return nil
	}
	content = bytes.NewReader(buffer.Bytes())
	return
}

// ToExcel 生成io.ReadSeeker  参数 titleList 为Excel表头，dataList 为数据
func ToExcel(dataList []interface{}) (content io.ReadSeeker) {
	// 生成一个新的文件
	file := xlsx.NewFile()

	// 添加sheet页
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		log.Printf("Failed to add sheet: %v", err)
		return nil
	}

	// 插入表头
	titleRow := sheet.AddRow()

	//获取表头
	objType := reflect.TypeOf(dataList[0])
	elem := objType.Elem()
	var titleList []string
	if elem.Kind() == reflect.Struct {
		for i := 1; i <= elem.NumField(); i++ {
			field := elem.Field(i - 1)
			if field.Name != "PageReq" {
				if field.Tag != "" && field.Tag.Get("description") != "" {
					titleList = append(titleList, i18n.T(field.Tag.Get("description")))
				} else {
					titleList = append(titleList, i18n.T(field.Name))
				}
			}
		}
	}

	// 创建表头样式
	headerStyle := xlsx.NewStyle()
	headerStyle.Font = *xlsx.NewFont(10, "Arial")
	headerStyle.Font.Color = "000000"
	headerStyle.Fill.PatternType = "solid"
	headerStyle.Fill.FgColor = "cfe2f3"
	headerStyle.Alignment.Horizontal = "center"
	headerStyle.Alignment.Vertical = "center"

	for _, v := range titleList {
		cell := titleRow.AddCell()
		cell.Value = v
		cell.SetStyle(headerStyle)
	}

	// 插入内容
	for _, v := range dataList {
		row := sheet.AddRow()
		writeStructToRow(row, v)
	}

	var buffer bytes.Buffer
	err = file.Write(&buffer)
	if err != nil {
		log.Printf("Failed to write file: %v", err)
		return nil
	}
	content = bytes.NewReader(buffer.Bytes())
	return
}

// writeStructToRow 将结构体写入Excel行
func writeStructToRow(row *xlsx.Row, v interface{}) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if val.Type().Field(i).Name != "PageReq" {
			cell := row.AddCell()
			switch field.Kind() {
			case reflect.String:
				cell.SetString(field.String())
			case reflect.Int, reflect.Int64:
				cell.SetInt64(field.Int())
			case reflect.Float64:
				cell.SetFloat(field.Float())
			case reflect.Bool:
				cell.SetBool(field.Bool())
			default:
				cell.SetString(fmt.Sprint(field.Interface()))
			}
		}
	}
}

func DownloadExcel(titleList []string, dataList []interface{}, filename ...string) (string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var fileName string
	if len(filename) > 0 && filename[0] != "" {
		fileName = filename[0]
	} else {
		curdate := time.Now().UnixNano()
		fileName = strconv.FormatInt(curdate, 10) + ".xlsx"
	}
	filePath := filepath.Join(curDir, "public", "upload", fileName)

	err = CreateFilePath(filePath)
	if err != nil {
		log.Printf("Failed to create file path: %v", err)
		return "", err
	}

	// 生成一个新的文件
	file := xlsx.NewFile()

	// 添加sheet页
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		return "", err
	}

	// 插入表头
	titleRow := sheet.AddRow()
	for _, v := range titleList {
		cell := titleRow.AddCell()
		cell.SetString(v)
	}

	// 插入内容
	for _, v := range dataList {
		row := sheet.AddRow()
		writeStructToRow(row, v)
	}

	// 保存文件
	err = file.Save(filePath)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

// CreateFilePath 创建路径
func CreateFilePath(filePath string) error {
	path := filepath.Dir(filePath)
	return os.MkdirAll(path, os.ModePerm)
}

// ReadExcelFile 读取EXCEL文件
func ReadExcelFile(r *http.Request, tableName ...string) (rows [][]string, err error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileName := r.FormValue("file_name")
	fileSuffix := strings.ToLower(filepath.Ext(fileName))

	if fileSuffix != ".xlsx" {
		return nil, errors.New("文件类型错误：仅支持.xlsx格式")
	}

	// 使用excelize读取Excel
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 获取sheet名
	sheetName := "Sheet1"
	if len(tableName) > 0 {
		sheetName = tableName[0]
	} else {
		sheetList := f.GetSheetList()
		if len(sheetList) > 0 {
			sheetName = sheetList[0]
		}
	}

	return f.GetRows(sheetName)
}
