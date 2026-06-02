package ai

import (
	"math/rand"
	"strings"
	"time"
)

type Assistant struct {
	enabled       bool
	mockResponses bool
}

func NewAssistant(enabled, mockResponses bool) *Assistant {
	return &Assistant{
		enabled:       enabled,
		mockResponses: mockResponses,
	}
}

// GenerateResponse generates a response for a user query
func (a *Assistant) GenerateResponse(query string) string {
	if !a.enabled {
		return "AI assistant is currently disabled."
	}

	if a.mockResponses {
		return a.getMockResponse(query)
	}

	// In production, this would call an actual AI API
	// For the diploma project, we use mock responses
	return a.getMockResponse(query)
}

// GetChatSummary generates a summary of chat messages
func (a *Assistant) GetChatSummary(messages []string) string {
	if !a.enabled {
		return "AI summary is currently disabled."
	}

	if len(messages) == 0 {
		return "No messages to summarize."
	}

	if a.mockResponses {
		return a.generateMockSummary(messages)
	}

	return a.generateMockSummary(messages)
}

// TranslateMessage translates a message to the target language
func (a *Assistant) TranslateMessage(text, targetLang string) string {
	if !a.enabled {
		return text
	}

	if a.mockResponses {
		return a.getMockTranslation(text, targetLang)
	}

	return a.getMockTranslation(text, targetLang)
}

func (a *Assistant) getMockResponse(query string) string {
	queryLower := strings.ToLower(query)
	rand.Seed(time.Now().UnixNano())

	responses := []string{
		"Я AI-ассистент Chelbo. Чем могу помочь?",
		"Отличный вопрос! Вот что я думаю по этому поводу...",
		"Спасибо за ваш запрос. Chelbo стремится предоставить лучший опыт общения!",
		"Интересно! Позвольте мне подумать над ответом.",
	}

	// Simple keyword-based responses
	if strings.Contains(queryLower, "привет") || strings.Contains(queryLower, "здравствуй") {
		return "Привет! Я AI-ассистент Chelbo. Рад помочь вам с любыми вопросами о мессенджере!"
	}
	if strings.Contains(queryLower, "chelbo") {
		return "Chelbo - это российский кроссплатформенный мессенджер с открытой архитектурой, работающий через Web и PWA без необходимости установки приложения."
	}
	if strings.Contains(queryLower, "как отправить") || strings.Contains(queryLower, "сообщение") {
		return "Чтобы отправить сообщение, выберите чат из списка слева, введите текст в поле ввода внизу и нажмите Enter или кнопку отправки."
	}
	if strings.Contains(queryLower, "группа") || strings.Contains(queryLower, "создать группу") {
		return "Чтобы создать группу, нажмите на кнопку 'Новый чат' и выберите 'Создать группу'. Затем укажите название группы и добавьте участников."
	}
	if strings.Contains(queryLower, "файл") || strings.Contains(queryLower, "отправить файл") {
		return "Для отправки файла нажмите на иконку скрепки в поле ввода сообщения. Вы можете отправлять изображения до 10 MB и документы до 50 MB."
	}
	if strings.Contains(queryLower, "безопасность") || strings.Contains(queryLower, "шифрование") {
		return "Chelbo использует TLS 1.3 для шифрования трафика, JWT токены для аутентификации и bcrypt для хэширования паролей. Все данные передаются по защищенным каналам."
	}

	return responses[rand.Intn(len(responses))]
}

func (a *Assistant) generateMockSummary(messages []string) string {
	if len(messages) == 0 {
		return "Чат пуст."
	}

	messageCount := len(messages)
	totalLength := 0
	for _, msg := range messages {
		totalLength += len(msg)
	}
	avgLength := totalLength / messageCount

	var topics []string
	if containsKeyword(messages, "привет", "здравствуй", "добрый") {
		topics = append(topics, "приветствия")
	}
	if containsKeyword(messages, "работа", "проект", "дело") {
		topics = append(topics, "рабочие вопросы")
	}
	if containsKeyword(messages, "погода", "сегодня", "завтра") {
		topics = append(topics, "обсуждение погоды")
	}
	if containsKeyword(messages, "спасибо", "отлично", "хорошо") {
		topics = append(topics, "положительные отзывы")
	}

	topicStr := "различные темы"
	if len(topics) > 0 {
		topicStr = strings.Join(topics, ", ")
	}

	return sprintf("В чате %d сообщений. Средняя длина сообщения: %d символов. Основные темы: %s.", messageCount, avgLength, topicStr)
}

func (a *Assistant) getMockTranslation(text, targetLang string) string {
	langNames := map[string]string{
		"ru": "русский",
		"en": "английский",
		"zh": "китайский",
		"es": "испанский",
		"fr": "французский",
		"de": "немецкий",
	}

	langName, ok := langNames[targetLang]
	if !ok {
		langName = targetLang
	}

	return sprintf("[Переведено на %s] %s", langName, text)
}

func containsKeyword(messages []string, keywords ...string) bool {
	for _, msg := range messages {
		msgLower := strings.ToLower(msg)
		for _, kw := range keywords {
			if strings.Contains(msgLower, kw) {
				return true
			}
		}
	}
	return false
}

func sprintf(format string, args ...interface{}) string {
	return strings.ReplaceAll(format, "%d", "%v")
}
