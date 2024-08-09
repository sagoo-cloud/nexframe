package i18n

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	setupTestEnvironment()
	code := m.Run()
	cleanupTestEnvironment()
	os.Exit(code)
}

func setupTestEnvironment() {
	createTestTranslationFile("en", map[string]string{
		"hello":   "Hello",
		"world":   "World",
		"welcome": "Welcome, %s!",
	})
	createTestTranslationFile("zh", map[string]string{
		"hello":   "你好",
		"world":   "世界",
		"welcome": "欢迎，%s！",
	})
}

func cleanupTestEnvironment() {
	os.RemoveAll("./test_translations")
}

func createTestTranslationFile(lang string, translations map[string]string) {
	data, _ := json.Marshal(translations)
	os.MkdirAll("./config/lang", 0755)
	ioutil.WriteFile(filepath.Join("./config/lang", lang+".json"), data, 0644)
}

func TestInitGlobal(t *testing.T) {
	err := InitGlobal("en", "./config/lang")
	if err != nil {
		t.Errorf("InitGlobal failed: %v", err)
	}
	if GetCurrentLang() != "en" {
		t.Errorf("Expected current language to be 'en', got '%s'", GetCurrentLang())
	}

	err = InitGlobal("invalid", "./config/lang")
	if err == nil {
		t.Error("Expected error for invalid language, got nil")
	}
}

func TestT(t *testing.T) {
	InitGlobal("en", "./config/lang")

	if T("hello") != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", T("hello"))
	}
	if T("world") != "World" {
		t.Errorf("Expected 'World', got '%s'", T("world"))
	}
	if T("unknown") != "unknown" {
		t.Errorf("Expected 'unknown', got '%s'", T("unknown"))
	}

	SetLang("zh")
	if T("hello") != "你好" {
		t.Errorf("Expected '你好', got '%s'", T("hello"))
	}
	if T("world") != "世界" {
		t.Errorf("Expected '世界', got '%s'", T("world"))
	}
}

func TestSetLang(t *testing.T) {
	InitGlobal("en", "./config/lang")

	err := SetLang("zh")
	if err != nil {
		t.Errorf("SetLang failed: %v", err)
	}
	if GetCurrentLang() != "zh" {
		t.Errorf("Expected current language to be 'zh', got '%s'", GetCurrentLang())
	}

	err = SetLang("invalid")
	if err == nil {
		t.Error("Expected error for invalid language, got nil")
	}
}

func TestFormatTranslation(t *testing.T) {
	InitGlobal("en", "./config/lang")

	result := FormatTranslation("welcome", "John")
	if result != "Welcome, John!" {
		t.Errorf("Expected 'Welcome, John!', got '%s'", result)
	}

	SetLang("zh")
	result = FormatTranslation("welcome", "Alice")
	if result != "欢迎，Alice！" {
		t.Errorf("Expected '欢迎，Alice！', got '%s'", result)
	}
}

func TestLoadTranslationsFromBytes(t *testing.T) {
	InitGlobal("en", "./config/lang")

	newTranslations := []byte(`{"new": "New Translation"}`)
	err := LoadTranslationsFromBytes("fr", newTranslations)
	if err != nil {
		t.Errorf("LoadTranslationsFromBytes failed: %v", err)
	}

	SetLang("fr")
	if T("new") != "New Translation" {
		t.Errorf("Expected 'New Translation', got '%s'", T("new"))
	}
}

func TestConcurrency(t *testing.T) {
	InitGlobal("en", "./config/lang")

	done := make(chan bool)
	go func() {
		for i := 0; i < 1000; i++ {
			T("hello")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			SetLang("zh")
			SetLang("en")
		}
		done <- true
	}()

	<-done
	<-done
}
