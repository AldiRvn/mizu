//coverage:ignore file
package msg

import (
	"context"
	"errors"
	"log/slog"
	"mime"
	"path/filepath"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type WhatsApp struct {
	client *whatsmeow.Client
}

func (w *WhatsApp) getJid(target any) (types.JID, error) {
	targetJid, ok := target.(types.JID)
	if !ok {
		err := errors.New("target is not JID")
		slog.Error("wa", "err", err)
		return types.EmptyJID, err
	}
	return targetJid, nil
}

func (w *WhatsApp) getMimeType(fileName string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}

func (w *WhatsApp) beforeSendMessage(target types.JID, media types.ChatPresenceMedia) {
	w.client.SendChatPresence(target, types.ChatPresenceComposing, media)
	time.Sleep(time.Second)
}

func (w *WhatsApp) afterSendMessage(target types.JID, media types.ChatPresenceMedia) {
	w.client.SendChatPresence(target, types.ChatPresencePaused, media)
}

func (w *WhatsApp) uploadFileToServerWa(ctx context.Context, content []byte, mediaType whatsmeow.MediaType) (uploaded whatsmeow.UploadResponse) {
	uploaded, err := w.client.Upload(
		ctx, content, mediaType)
	if err != nil {
		slog.Error("wa upload file", "err", err)
		return
	}
	return
}

func (w *WhatsApp) SendText(ctx context.Context, target any, text string) error {
	media := types.ChatPresenceMediaText
	targetJid, err := w.getJid(target)
	if err != nil {
		return err
	}

	w.beforeSendMessage(targetJid, media)
	defer w.afterSendMessage(targetJid, media)

	w.client.SendMessage(ctx, targetJid, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &text,
		},
	})

	return nil
}

func (w *WhatsApp) SendDocument(ctx context.Context, target any, document []byte, fileName string) error {
	media := types.ChatPresenceMediaText
	targetJid, err := w.getJid(target)
	if err != nil {
		return err
	}

	w.beforeSendMessage(targetJid, media)
	defer w.afterSendMessage(targetJid, media)

	mimeType := w.getMimeType(fileName)
	uploaded := w.uploadFileToServerWa(ctx, document, whatsmeow.MediaDocument)

	w.client.SendMessage(ctx, targetJid, &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			ContextInfo:   &waE2E.ContextInfo{},
			Title:         proto.String(fileName), // Nama yang keliatan di WA
			FileName:      proto.String(fileName),
			Mimetype:      proto.String(mimeType),
			URL:           proto.String(uploaded.URL),
			DirectPath:    &uploaded.DirectPath,
			FileSHA256:    uploaded.FileSHA256,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileLength:    proto.Uint64(uint64(len(document))),
			MediaKey:      uploaded.MediaKey,
		},
	})
	return nil
}

func (w *WhatsApp) SendImage(ctx context.Context, target any, image []byte, fileName string) error {
	media := types.ChatPresenceMediaText
	targetJid, err := w.getJid(target)
	if err != nil {
		return err
	}

	w.beforeSendMessage(targetJid, media)
	defer w.afterSendMessage(targetJid, media)

	mimeType := w.getMimeType(fileName)
	uploaded := w.uploadFileToServerWa(ctx, image, whatsmeow.MediaImage)

	w.client.SendMessage(ctx, targetJid, &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			ContextInfo: &waE2E.ContextInfo{},
			Mimetype:    proto.String(mimeType),

			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	})
	return nil
}
