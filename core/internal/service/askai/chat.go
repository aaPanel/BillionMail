package askai

import (
	"billionmail-core/internal/service/public"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	CHAT_CONFIG_PATH = "../conf/chat"
)

type TempChat struct {
	ChatId      string `json:"chatId"`       // Unique identifier for the chat
	Domain      string `json:"domain"`       // Domain associated with the chat
	AddType     int    `json:"add_type"`     // Type of addition (e.g., 0 for new chat, 1 for existing chat)
	TempName    string `json:"temp_name"`    // Temporary name for the chat or knowledge base file
	HtmlContent string `json:"html_content"` // HTML content of the chat or knowledge base file
	DragData    string `json:"drag_data"`    // Data for drag-and-drop functionality
	CreateTime  int64  `json:"create_time"`  // Creation time of the chat
	UpdateTime  int64  `json:"update_time"`  // Last update time of the chat
}

type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Number of tokens in the prompt
	CompletionTokens int `json:"completion_tokens"` // Number of tokens in the completion
	TotalTokens      int `json:"total_tokens"`      // Total number of tokens used
}

type Message struct {
	MessageId     string    `json:"messageId"`      // Unique identifier for the message
	ChatId        string    `json:"chatId"`         // Unique identifier for the chat
	Ppid          string    `json:"ppid"`           // Parent message ID (if applicable)
	Content       string    `json:"content"`        // Content of the message
	HtmlContent   string    `json:"html_content"`   // HTML content of the message
	Role          string    `json:"role"`           // Role of the message sender (e.g., "user", "assistant","system","tool")
	CreateTime    int64     `json:"create_time"`    // Creation time of the message
	EndTime       int64     `json:"end_time"`       // End time of the message (if applicable)
	Reasoning     string    `json:"reasoning"`      // Reasoning or explanation for the message
	SupplierName  string    `json:"supplier_name"`  // Name of the AI model supplier
	ModelId       string    `json:"model_id"`       // Identifier for the AI model used in the message
	FinishReason  string    `json:"finish_reason"`  // Reason for finishing the message (e.g., "stop", "length")
	Usage         ChatUsage `json:"usage"`          // Usage statistics for the message
	TimeConsuming int64     `json:"time_consuming"` // Time taken to process the message in milliseconds
}

type File struct {
	Fid        string `json:"fid"`         // Unique identifier for the file
	Content    string `json:"content"`     // Content of the file
	Title      string `json:"title"`       // Title of the file
	Size       int64  `json:"size"`        // Size of the file in bytes
	UpdateTime int64  `json:"update_time"` // Last update time of the file
}

type ChatInfo struct {
	ChatId       string    `json:"chatId"`       // Unique identifier for the chat
	Domain       string    `json:"domain"`       // Domain associated with the chat
	Prompt       string    `json:"prompt"`       // Prompt or question for the chat
	Title        string    `json:"title"`        // Title of the chat
	SupplierName string    `json:"supplierName"` // Name of the AI model supplier
	ModelId      string    `json:"modelId"`      // Identifier for the AI model used in the chat
	Messages     []Message `json:"messages"`     // List of messages in the chat
	Files        []File    `json:"files"`        // List of files associated with the chat
	CreateTime   int64     `json:"create_time"`  // Creation time of the chat
	UpdateTime   int64     `json:"update_time"`  // Last update time of the chat
}

var ChatStatus = map[string]bool{}

func SaveTempChat(chatId string, chatInfo *TempChat) error {
	// Implementation for saving a temporary chat
	// This function should handle the logic for saving the temporary chat information
	tempChatPath := CHAT_CONFIG_PATH + "/" + chatId
	if !public.FileExists(tempChatPath) {
		os.MkdirAll(tempChatPath, os.ModePerm)
	}
	jsonData, err := json.Marshal(chatInfo)
	if err != nil {
		return err
	}
	tempChatFile := tempChatPath + "/config.json"
	os.WriteFile(tempChatFile, jsonData, 0644)
	return nil
}

func GetTempChat(chatId string) (*TempChat, error) {
	// Implementation for retrieving a temporary chat
	// This function should handle the logic for loading the temporary chat information
	tempChatPath := CHAT_CONFIG_PATH + "/" + chatId + "/config.json"
	if !public.FileExists(tempChatPath) {
		return nil, os.ErrNotExist
	}
	data, err := os.ReadFile(tempChatPath)
	if err != nil {
		return nil, err
	}
	var tempChat TempChat
	err = json.Unmarshal(data, &tempChat)
	if err != nil {
		return nil, err
	}
	return &tempChat, nil
}

func SaveChat(chatId string, chatInfo *ChatInfo) error {
	// Implementation for saving chat information
	// This function should handle the logic for saving the chat information
	chatPath := CHAT_CONFIG_PATH + "/" + chatId
	if !public.FileExists(chatPath) {
		os.MkdirAll(chatPath, os.ModePerm)
	}
	jsonData, err := json.Marshal(chatInfo)
	if err != nil {
		return err
	}
	chatFile := chatPath + "/info.json"
	return os.WriteFile(chatFile, jsonData, 0644)
}

func GetChat(chatId string) (*ChatInfo, error) {
	// Implementation for retrieving chat information
	// This function should handle the logic for loading the chat information
	chatPath := CHAT_CONFIG_PATH + "/" + chatId + "/info.json"
	if !public.FileExists(chatPath) {
		return nil, os.ErrNotExist
	}
	data, err := os.ReadFile(chatPath)
	if err != nil {
		return nil, err
	}
	var chatInfo ChatInfo
	err = json.Unmarshal(data, &chatInfo)
	if err != nil {
		return nil, err
	}
	return &chatInfo, nil
}

func GetMessages(chatId string) []Message {
	messageFile := CHAT_CONFIG_PATH + "/" + chatId + "/message.json"
	if !public.FileExists(messageFile) {
		return []Message{} // Return an empty slice if the file does not exist
	}
	data, err := os.ReadFile(messageFile)
	if err != nil {
		return []Message{} // Return an empty slice if there's an error reading the file
	}
	var messages []Message
	err = json.Unmarshal(data, &messages)
	if err != nil {
		return []Message{} // Return an empty slice if there's an error unmarshalling the data
	}
	return messages
}

func CreateChat(Domain string, AddType int, TempName string) (string, error) {
	// Implementation for creating a chat
	// This function should handle the logic for creating a chat based on the provided parameters
	chatId := GetUUID()
	pData := TempChat{
		ChatId:      chatId,
		Domain:      Domain,
		AddType:     AddType,
		TempName:    TempName,
		HtmlContent: "",
		DragData:    "",
		CreateTime:  public.GetNowTime(),
		UpdateTime:  public.GetNowTime(),
	}

	err := SaveTempChat(chatId, &pData)
	if err != nil {
		return "", err
	}

	// Initialize chat information with default values
	// This function should handle the logic for initializing a chat with default values
	chatInfo := ChatInfo{
		ChatId:       chatId,
		Domain:       Domain,
		Prompt:       "",
		Title:        TempName,
		SupplierName: "",
		ModelId:      "",
		Messages:     []Message{},
		Files:        []File{},
		CreateTime:   public.GetNowTime(),
		UpdateTime:   public.GetNowTime(),
	}
	err = SaveChat(chatId, &chatInfo)
	if err != nil {
		return "", err
	}

	return chatId, nil
}

func GetChatList() ([]TempChat, error) {
	var chatList []TempChat
	// Implementation for retrieving the chat list
	// This function should return the list of chats and any error encountered
	if !public.FileExists(CHAT_CONFIG_PATH) {
		return chatList, nil
	}
	files, err := os.ReadDir(CHAT_CONFIG_PATH)
	if err != nil {
		return chatList, err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		chatId := file.Name()
		tempChat, err := GetTempChat(chatId)
		if err != nil {
			continue // Skip if there's an error retrieving the temporary chat
		}

		chatList = append(chatList, *tempChat)
	}

	return chatList, nil
}

// GetChatInfo retrieves the chat information for a given chat ID
// This function should handle the logic for loading the chat information and return the ChatInfo struct and any error encountered
func GetChatInfo(chatId string) (*ChatInfo, error) {
	chatInfo, err := GetChat(chatId)
	if err != nil {
		return nil, err
	}
	chatInfo.Messages = GetMessages(chatId)
	return chatInfo, nil
}

func Chat(ctx context.Context, chatId string, supplierName string, modelId string, content string, isText bool) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("Content cannot be empty")
	}

	supplierInfo, err := GetSupplierConfig(supplierName)
	if err != nil {
		return errors.New("Supplier not found")
	}

	modelInfo := GetModelInfo(supplierName, modelId)
	if modelInfo == nil {
		return errors.New("Model not found")
	}

	// Check if the chat is already active
	chatInfo, err := GetChat(chatId)
	if err != nil {
		return errors.New("Chat not found")
	}
	// set the chat information
	chatInfo.SupplierName = supplierName
	chatInfo.ModelId = modelId
	err = SaveChat(chatId, chatInfo)
	if err != nil {
		return errors.New("Failed to save chat")
	}

	ChatStatus[chatId] = true // Set chat status to active
	defer func() {
		delete(ChatStatus, chatId) // Ensure chat status is removed after processing
	}()

	aiObj := NewOpenAI(ctx, supplierInfo.ApiKey, supplierInfo.BaseUrl, modelId, chatId, supplierName, modelInfo.MaxTokens)
	if aiObj == nil {
		return errors.New("failed to initialize AI client")
	}
	aiObj.GetClient()
	return aiObj.Chat(content, isText)
}

// RemoveChat removes a chat by its ID
// This function should handle the logic for removing a chat based on the provided chat ID
func RemoveChat(chatId string) error {
	chatPath := CHAT_CONFIG_PATH + "/" + chatId
	if !public.FileExists(chatPath) {
		return os.ErrNotExist
	}
	err := os.RemoveAll(chatPath)
	if err != nil {
		return err
	}
	return nil
}

func Stop(chatId string) error {
	// Implementation for stopping a chat
	// This function should handle the logic for stopping a chat based on the provided parameters
	if _, exists := ChatStatus[chatId]; !exists {
		return errors.New("Chat not found or already stopped")
	}
	ChatStatus[chatId] = false // Set chat status to inactive
	return nil
}

// GetLastUsage retrieves the last usage statistics for a chat by its ID
// This function should handle the logic for loading the last usage statistics of a chat based on the provided chat ID
func GetLastUsage(chatId string) ChatUsage {
	messages := GetMessages(chatId)
	if len(messages) == 0 {
		return ChatUsage{}
	}
	lastMessage := messages[len(messages)-1]
	return lastMessage.Usage
}

// GetHtml retrieves the HTML content of a chat by its ID
// This function should handle the logic for loading the HTML content of a chat based on the provided chat ID
func GetHtml(chatId string) (string, error) {
	var htmlContent string
	filename := CHAT_CONFIG_PATH + "/" + chatId + "/code.html"
	if !public.FileExists(filename) {
		messages := GetMessages(chatId)
		messagesLen := len(messages)
		if messagesLen == 0 {
			return "", nil
		}
		message := messages[messagesLen-1]
		if message.HtmlContent == "" {
			if strings.Contains(message.Content, `<!DOCTYPE html>`) && strings.Contains(message.Content, "</html>") {
				// 截取 HTML 内容
				startIndex := strings.Index(message.Content, "<!DOCTYPE html>")
				if startIndex == -1 {
					return "", nil
				}
				endIndex := strings.LastIndex(message.Content, "</html>")
				if endIndex == -1 {
					return "", nil
				}
				htmlContent = message.Content[startIndex : endIndex+len("</html>")]
				return htmlContent, nil
			}
		}
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	htmlContent = string(data)
	return htmlContent, nil
}

// ModifyHtml modifies the HTML content of a chat by its ID
// This function should handle the logic for modifying the HTML content of a chat based on the provided chat ID and content
func ModifyHtml(chatId string, content string) error {
	filename := CHAT_CONFIG_PATH + "/" + chatId + "/code.html"
	// if !public.FileExists(filename) {
	// 	return errors.New("HTML file not found")
	// }
	g.DB().Model("email_templates").Where("chat_id", chatId).Update(g.Map{
		"content": content, // Update the HTML content in the database
	})

	return os.WriteFile(filename, []byte(content), os.ModePerm)
}

func CopyChat(chatId string, domain string) (string, string, error) {
	// Implementation for copying a chat
	// This function should handle the logic for copying a chat based on the provided chat ID and domain

	chatInfo, err := GetChat(chatId)
	if err != nil {
		return "", "", err
	}

	newChatId := GetUUID()
	newTempChat := TempChat{
		ChatId:     newChatId,
		Domain:     domain,
		AddType:    99,
		CreateTime: public.GetNowTime(),
		UpdateTime: public.GetNowTime(),
	}
	err = SaveTempChat(newChatId, &newTempChat)

	if err != nil {
		return "", "", err
	}

	newChatInfo := ChatInfo{
		ChatId:       newChatId,
		Prompt:       chatInfo.Prompt,
		Title:        chatInfo.Title,
		SupplierName: chatInfo.SupplierName,
		ModelId:      chatInfo.ModelId,
		Messages:     []Message{},
		Files:        []File{},
		CreateTime:   public.GetNowTime(),
		UpdateTime:   public.GetNowTime(),
	}

	err = SaveChat(newChatId, &newChatInfo)
	if err != nil {
		return "", "", err
	}

	oldChatId := chatInfo.ChatId
	oldCodeFile := CHAT_CONFIG_PATH + "/" + oldChatId + "/code.html"
	newCodeFile := CHAT_CONFIG_PATH + "/" + newChatId + "/code.html"
	oldCode := ""
	if public.FileExists(oldCodeFile) {
		oldCodeBytes, err := os.ReadFile(oldCodeFile)
		oldCode = string(oldCodeBytes)
		if err != nil {
			return "", oldCode, err
		}
		err = os.WriteFile(newCodeFile, oldCodeBytes, os.ModePerm)
		if err != nil {
			return "", oldCode, err
		}
	}

	return newChatId, oldCode, nil
}
