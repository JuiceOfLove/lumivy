import React, { useRef, useMemo } from "react";
import { API_URL } from "../../../../../http";
import { IChatMessage } from "../../../../../types/chat";
import styles from "./OperatorMessageList.module.css";

interface Props {
  messages: IChatMessage[];
  currentUserId: number;
  operatorId: number;
  users: { id: number; name: string }[];
  onAction: (a: "reply" | "delete", m: IChatMessage) => void;
}

const FILES_ORIGIN = API_URL.replace(/\/api\/?$/, "");

const OperatorMessageList: React.FC<Props> = ({
  messages,
  currentUserId,
  operatorId,
  users,
  onAction,
}) => {
  const nameOf = useMemo(() => {
    const m = new Map<number, string>();
    users.forEach(u => m.set(u.id, u.name));
    return (id: number) => m.get(id) ?? `ID${id}`;
  }, [users]);

  const refs = useRef<Record<number, HTMLDivElement | null>>({});

  const scrollTo = (id: number) =>
    refs.current[id]?.scrollIntoView({ behavior: "smooth", block: "center" });

  const context = (msg: IChatMessage) => (e: React.MouseEvent) => {
    e.preventDefault();
    const menu = document.createElement("div");
    menu.className = styles.menu;
    menu.style.left = `${e.clientX}px`;
    menu.style.top = `${e.clientY}px`;

    const add = (text: string, cb: () => void) => {
      const item = document.createElement("div");
      item.className = styles.menuItem;
      item.textContent = text;
      item.onclick = () => { cb(); menu.remove(); };
      menu.appendChild(item);
    };
    add("Ответить", () => onAction("reply", msg));
    add("Удалить", () => onAction("delete", msg));
    document.body.appendChild(menu);
    window.addEventListener("click", () => menu.remove(), { once: true });
  };

  return (
    <>
      {messages.filter(m => !m.deleted_at).map(m => {
        const mine = m.user_id === currentUserId;
        return (
          <div
            key={m.id}
            ref={el => refs.current[m.id] = el}
            className={`${styles.msg} ${mine ? styles.mine : styles.theirs}`}
            onContextMenu={context(m)}
          >
            <div className={styles.author}>
              {nameOf(m.user_id)}
              {(!mine && m.user_id === operatorId) && (
                <span className={styles.operatorLabel}>operator</span>
              )}
            </div>

            {m.reply_to_id && (
              <div className={styles.reply} onClick={() => scrollTo(m.reply_to_id!)}>
                <span className={styles.replyLine} />
                <span className={styles.replyText}>
                  {messages.find(x => x.id === m.reply_to_id)?.content || "…"}
                </span>
              </div>
            )}

            {m.media_url && (
              <div className={styles.media}>
                <img src={FILES_ORIGIN + m.media_url} alt="attachment" />
              </div>
            )}

            {m.content && <div className={styles.text}>{m.content}</div>}

            <div className={styles.meta}>
              {new Date(m.created_at).toLocaleTimeString()}
            </div>
          </div>
        );
      })}
    </>
  );
};

export default OperatorMessageList;