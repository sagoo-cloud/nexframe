// Package i18n 提供了一个简单的国际化（i18n）支持库
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"golang.org/x/text/language"
)

var (
	// translations 存储所有语言的翻译
	translations sync.Map
	// currentLang 存储当前使用的语言
	currentLang atomic.Value
	// defaultLang 定义默认语言
	defaultLang = language.English
	// translationDir 存储翻译文件的目录
	translationDir string
	// fileLock 用于确保文件操作的线程安全
	fileLock sync.Mutex
)

// translator 结构体用于存储单个语言的翻译
type translator struct {
	translations map[string]string
}

// InitGlobal 初始化全局 i18n 实例
// lang: 初始语言
// customDir: 可选的自定义翻译文件目录
func InitGlobal(lang string, customDir ...string) error {
	parsedLang, err := language.Parse(lang)
	if err != nil {
		return fmt.Errorf("无效的语言: %w", err)
	}

	currentLang.Store(parsedLang)

	// 使用自定义目录（如果提供），否则使用默认目录
	if len(customDir) > 0 {
		translationDir = customDir[0]
	} else {
		translationDir = "config/lang"
	}

	return loadLanguage(parsedLang)
}

// loadLanguage 加载特定语言的翻译
func loadLanguage(lang language.Tag) error {
	filename := getTranslationFilePath(lang)
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，初始化为空翻译
			translations.Store(lang, &translator{translations: make(map[string]string)})
			return nil
		}
		return fmt.Errorf("读取翻译文件 %s 时出错: %w", filename, err)
	}

	var langTranslations map[string]string
	err = json.Unmarshal(data, &langTranslations)
	if err != nil {
		return fmt.Errorf("解析 %s 的翻译时出错: %w", lang, err)
	}

	translations.Store(lang, &translator{translations: langTranslations})
	return nil
}

// getTranslationFilePath 返回翻译文件的完整路径
func getTranslationFilePath(lang language.Tag) string {
	return filepath.Join(translationDir, lang.String()+".json")
}

// T 翻译给定的键
func T(key string) string {
	lang := currentLang.Load().(language.Tag)
	if t, ok := translations.Load(lang); ok {
		if translation, exists := t.(*translator).translations[key]; exists {
			return translation
		}
	}

	// 如果在当前语言中找不到翻译，尝试使用默认语言
	if lang != defaultLang {
		if t, ok := translations.Load(defaultLang); ok {
			if translation, exists := t.(*translator).translations[key]; exists {
				return translation
			}
		}
	}

	return key
}

// SetLang 更改当前语言
func SetLang(lang string) error {
	parsedLang, err := language.Parse(lang)
	if err != nil {
		return fmt.Errorf("无效的语言: %w", err)
	}

	if _, ok := translations.Load(parsedLang); !ok {
		err := loadLanguage(parsedLang)
		if err != nil {
			return fmt.Errorf("加载语言 %s 时出错: %w", lang, err)
		}
	}

	currentLang.Store(parsedLang)
	return nil
}

// GetCurrentLang 返回当前语言
func GetCurrentLang() string {
	return currentLang.Load().(language.Tag).String()
}

// AddTranslation 为当前语言添加或更新翻译，并保存到文件
func AddTranslation(key, translation string) error {
	lang := currentLang.Load().(language.Tag)

	t, ok := translations.Load(lang)
	if !ok {
		t = &translator{translations: make(map[string]string)}
		translations.Store(lang, t)
	}

	t.(*translator).translations[key] = translation

	return saveTranslationsToFile(lang)
}

// saveTranslationsToFile 将特定语言的翻译保存到文件
func saveTranslationsToFile(lang language.Tag) error {
	fileLock.Lock()
	defer fileLock.Unlock()

	t, ok := translations.Load(lang)
	if !ok {
		return fmt.Errorf("未找到语言 %s 的翻译", lang)
	}

	data, err := json.MarshalIndent(t.(*translator).translations, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化翻译时出错: %w", err)
	}

	filename := getTranslationFilePath(lang)
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("将翻译写入文件 %s 时出错: %w", filename, err)
	}

	return nil
}

// FormatTranslation 格式化带参数的翻译
func FormatTranslation(key string, args ...interface{}) string {
	translation := T(key)
	return fmt.Sprintf(translation, args...)
}

// LoadTranslationsFromBytes 从字节数组加载翻译
func LoadTranslationsFromBytes(lang string, data []byte) error {
	parsedLang, err := language.Parse(lang)
	if err != nil {
		return fmt.Errorf("无效的语言: %w", err)
	}

	var langTranslations map[string]string
	err = json.Unmarshal(data, &langTranslations)
	if err != nil {
		return fmt.Errorf("解析 %s 的翻译时出错: %w", lang, err)
	}

	translations.Store(parsedLang, &translator{translations: langTranslations})
	return nil
}
