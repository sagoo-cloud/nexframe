package utils

import (
	"bytes"
	"errors"
	"github.com/sagoo-cloud/nexframe/i18n"
	"github.com/tealeg/xlsx"
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

	for sheet, dataList := range data {
		// 添加sheet页
		sheet, _ := file.AddSheet(sheet)
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

		for _, v := range titleList {
			cell := titleRow.AddCell()
			cell.Value = v
			//表头字体颜色
			cell.GetStyle().Font.Color = "000000"
			cell.GetStyle().Fill.BgColor = "cfe2f3"
			//居中显示
			cell.GetStyle().Alignment.Horizontal = "center"
			cell.GetStyle().Alignment.Vertical = "center"
		}
		// 插入内容
		for _, v := range dataList {
			row := sheet.AddRow()
			row.WriteStruct(v, -1)
		}
	}

	var buffer bytes.Buffer
	_ = file.Write(&buffer)
	content = bytes.NewReader(buffer.Bytes())
	return
}

// ToExcel 生成io.ReadSeeker  参数 titleList 为Excel表头，dataList 为数据
func ToExcel(dataList []interface{}) (content io.ReadSeeker) {
	// 生成一个新的文件
	file := xlsx.NewFile()
	// 添加sheet页
	sheet, _ := file.AddSheet("Sheet1")
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

	for _, v := range titleList {
		cell := titleRow.AddCell()
		cell.Value = v
		//表头字体颜色
		cell.GetStyle().Font.Color = "000000"
		cell.GetStyle().Fill.BgColor = "cfe2f3"
		//居中显示
		cell.GetStyle().Alignment.Horizontal = "center"
		cell.GetStyle().Alignment.Vertical = "center"
	}
	// 插入内容
	for _, v := range dataList {
		row := sheet.AddRow()
		row.WriteStruct(v, -1)
	}

	var buffer bytes.Buffer
	_ = file.Write(&buffer)
	content = bytes.NewReader(buffer.Bytes())
	return
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
		fileName = strconv.FormatInt(curdate, 10) + ".xls"
	}
	filePath := curDir + "/public/upload/" + fileName

	err = CreateFilePath(filePath)
	if err != nil {
		log.Printf("%s", err.Error())
		return "", err
	}

	// 生成一个新的文件
	file := xlsx.NewFile()
	// 添加sheet页
	sheet, _ := file.AddSheet("Sheet1")
	// 插入表头
	titleRow := sheet.AddRow()
	for _, v := range titleList {
		cell := titleRow.AddCell()
		cell.Value = v
	}
	// 插入内容
	for _, v := range dataList {
		row := sheet.AddRow()
		row.WriteStruct(v, -1)
	}

	// 在提供的路径中将文件保存到xlsx文件
	err = file.Save(filePath)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

// CreateFilePath  创建路径
func CreateFilePath(filePath string) error {
	// 路径不存在创建路径
	path, _ := filepath.Split(filePath) // 获取路径
	_, err := os.Stat(path)             // 检查路径状态，不存在创建
	if err != nil || os.IsExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
	}
	return err
}

// ReadExcelFile 读取EXCEL文件
func ReadExcelFile(r *http.Request, tableName ...string) (rows [][]string, err error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileName := r.FormValue("file_name") // 获取文件名字
	// 获取文件后缀
	var fileSuffix = fileName[strings.LastIndex(fileName, ".")+1:]

	// 判断文件后缀是否为.xlsx
	if !strings.EqualFold(fileSuffix, "xlsx") {
		return nil, errors.New("文件类型错误")
	}

	// 使用excelize从文件流中打开Excel文件
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}

	// 默认读取第一个sheet
	firstSheet := ""
	if len(tableName) > 0 {
		firstSheet = tableName[0]
	} else {
		firstSheet = f.GetSheetName(0)
	}

	rows, err = f.GetRows(firstSheet)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
