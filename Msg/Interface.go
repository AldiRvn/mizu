package msg

import (
	"context"
	"log/slog"
)

type Messager interface {
	SendText(ctx context.Context, target any, text string) error
	SendDocument(ctx context.Context, target any, content []byte, fileName string) error
	SendImage(ctx context.Context, target any, content []byte, fileName string) error
}

type Message struct {
	messager Messager
}

func NewMessage(m Messager) *Message {
	return &Message{messager: m}
}

func (m *Message) SendText(ctx context.Context, target any, text string) error {
	slog.Debug("msg send", "text", text)
	return m.messager.SendText(ctx, target, text)
}

func (m *Message) SendDocument(ctx context.Context, target any, document []byte, fileName string) error {
	slog.Debug("msg send", "documentLength", len(document), "fileName", fileName)
	return m.messager.SendDocument(ctx, target, document, fileName)
}

func (m *Message) SendImage(ctx context.Context, target any, image []byte, fileName string) error {
	slog.Debug("msg send", "imageLength", len(image), "fileName", fileName)
	return m.messager.SendImage(ctx, target, image, fileName)
}
