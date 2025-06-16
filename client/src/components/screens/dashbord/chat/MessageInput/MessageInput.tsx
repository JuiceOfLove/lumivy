import React, { useState, ChangeEvent, FormEvent, useEffect } from 'react';
import ChatService from '../../../../../services/ChatService';
import { IChatMessage } from '../../../../../types/chat';
import styles from './MessageInput.module.css';

interface Props {
  replyTo: IChatMessage | null;
  onCancelReply: () => void;
}

const MessageInput: React.FC<Props> = ({ replyTo, onCancelReply }) => {
  const [text, setText] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [previewURL, setPreviewURL] = useState<string | null>(null);
  const [caption, setCaption] = useState('');

  useEffect(() => {
    if (!file) return;
    const url = URL.createObjectURL(file);
    setPreviewURL(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  const sendPlain = (e: FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;
    ChatService.send(text.trim(), replyTo?.id);
    setText('');
    onCancelReply();
  };

  const sendWithPhoto = () => {
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      ChatService.send(caption.trim(), replyTo?.id, reader.result as string);
      closePreview();
    };
    reader.readAsDataURL(file);
  };

  const closePreview = () => {
    setFile(null);
    setPreviewURL(null);
    setCaption('');
  };

  return (
    <>
      {file && previewURL && (
        <div className={styles.overlay} onClick={closePreview}>
          <form className={styles.modal} onClick={e => e.stopPropagation()}>
            <button type="button" className={styles.close} onClick={closePreview}>×</button>
            <h3 className={styles.title}>Send Photo</h3>

            <img src={previewURL} className={styles.preview} alt="preview" />

            <input
              className={styles.caption}
              placeholder="Добавьте комментарий..."
              value={caption}
              onChange={e => setCaption(e.target.value)}
            />

            <button type="button" className={styles.sendBtn} onClick={sendWithPhoto}>
              Отправить
            </button>
          </form>
        </div>
      )}

      <form className={styles.box} onSubmit={sendPlain}>
        {replyTo && (
          <div className={styles.replyBar}>
            ↺ {replyTo.content?.slice(0, 50)}
            <button type="button" onClick={onCancelReply}>×</button>
          </div>
        )}

        <input
          type="text"
          placeholder="Сообщение"
          value={text}
          onChange={e => setText(e.target.value)}
        />

        <label className={styles.clip}>
          📎
          <input
            type="file"
            accept="image/*"
            hidden
            onChange={(e: ChangeEvent<HTMLInputElement>) => {
              const f = e.target.files?.[0];
              if (f) setFile(f);
            }}
          />
        </label>

        <button type="submit" disabled={!text.trim()}>➤</button>
      </form>
    </>
  );
};

export default MessageInput;