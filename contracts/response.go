package contracts

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// JsonRes 数据返回通用JSON数据结构
type JsonRes struct {
	Code    int         `json:"code"`    // 错误码((0:成功, 1:失败, >1:错误码))
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 返回数据(业务接口定义具体数据结构)
}

// JSONResp 返回标准JSON数据。
func JSONResp(w http.ResponseWriter, code int, message string, data ...interface{}) {
	var responseData interface{}
	if len(data) > 0 {
		responseData = data[0]
	} else {
		responseData = map[string]interface{}{}
	}
	jsonRes := JsonRes{
		Code:    code,
		Message: message,
		Data:    responseData,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonRes)
}

// JsonExit 返回标准JSON数据并退出当前HTTP执行函数。
func JsonExit(w http.ResponseWriter, code int, message string, data ...interface{}) {
	JSONResp(w, code, message, data...)
}

// JsonRedirect 返回标准JSON数据引导客户端跳转。
func JsonRedirect(w http.ResponseWriter, code int, message, redirect string, data ...interface{}) {
	responseData := interface{}(nil)
	if len(data) > 0 {
		responseData = data[0]
	}
	jsonRes := JsonRes{
		Code:    code,
		Message: message,
		Data:    responseData,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonRes)
}

// JsonRedirectExit 返回标准JSON数据引导客户端跳转，并退出当前HTTP执行函数。
func JsonRedirectExit(w http.ResponseWriter, code int, message, redirect string, data ...interface{}) {
	JsonRedirect(w, code, message, redirect, data...)
}

// ToXls 向前端返回Excel文件 参数 content 为上面生成的io.ReadSeeker， fileTag 为返回前端的文件名
func ToXls(w http.ResponseWriter, content io.ReadSeeker, fileTag string) {
	fileName := fmt.Sprintf("%s%s%s.xlsx", time.Now().Format("20060102150405"), `-`, fileTag)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	http.ServeContent(w, nil, fileName, time.Now(), content)
}

// ToPlainText 输出流
func ToPlainText(w http.ResponseWriter, content []byte, fileName string) {
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
	w.Write(content)
}

// ToJsonFIle 向前端返回文件 参数 content 为上面生成的io.ReadSeeker， fileTag 为返回前端的文件名
func ToJsonFIle(w http.ResponseWriter, content io.ReadSeeker, fileTag string) {
	fileName := fmt.Sprintf("%s%s%s.json", time.Now().Format("20060102150405"), `-`, fileTag)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, nil, fileName, time.Now(), content)
}
