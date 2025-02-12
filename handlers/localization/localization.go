package localization

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"tosdrgo/internal/logger"
)

var (
	translations = make(map[string]map[string]string)
	mutex        sync.RWMutex
)

func LoadTranslations(lang string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := translations[lang]; exists {
		return nil
	}

	file, err := os.ReadFile("locales/" + lang + ".json")
	if err != nil {
		// If language file doesn't exist, load English as fallback
		logger.LogWarn("Language file for %s not found, loading English as fallback", lang)
		if lang != "en" {
			file, err = os.ReadFile("locales/en.json")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var langData map[string]string
	if err := json.Unmarshal(file, &langData); err != nil {
		return err
	}

	translations[lang] = langData
	logger.LogDebug("Loaded translations for %s: %+v", lang, langData)
	return nil
}

func Get(lang, key string) string {
	mutex.RLock()
	defer mutex.RUnlock()

	if trans, ok := translations[lang]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}

	// Fallback to English
	if lang != "en" {
		if enTrans, ok := translations["en"]; ok {
			if val, ok := enTrans[key]; ok {
				return val
			}
		}
	}

	return key
}

func GetFormatted(lang, key string, args ...interface{}) string {
	mutex.RLock()
	defer mutex.RUnlock()

	if trans, ok := translations[lang]; ok {
		if val, ok := trans[key]; ok {
			return fmt.Sprintf(val, args...)
		}
	}

	// Fallback to English
	if lang != "en" {
		if enTrans, ok := translations["en"]; ok {
			if val, ok := enTrans[key]; ok {
				return fmt.Sprintf(val, args...)
			}
		}
	}

	return key
}
