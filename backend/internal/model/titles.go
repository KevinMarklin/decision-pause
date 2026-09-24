package model

// ScenarioTitle — заголовок и эмодзи по ключу сценария (в БД не храним).
func ScenarioTitle(key string) (title, emoji string) {
	switch key {
	case "expected":
		return "Ожидаемый", "🟢"
	case "moderate":
		return "Умеренно негативный", "🟡"
	case "negative":
		return "Негативный", "🔴"
	default:
		return key, ""
	}
}
