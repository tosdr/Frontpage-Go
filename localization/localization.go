package localization

import (
	"encoding/json"
	"os"
	"sync"
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
