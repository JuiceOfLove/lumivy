import React, { useEffect, useState, useRef, useContext } from "react";
import { useParams, useNavigate } from "react-router";
import { Context } from "../../../../../main";
import SupportService, { ITicketInfo } from "../../../../../services/SupportService";
import SupportSocketService from "../../../../../services/SupportSocketService";
import { IChatMessage } from "../../../../../types/chat";
import styles from "./TicketChat.module.css";
import MessageList from "../../chat/MessageList/MessageList";
import OperatorMessageList from "../OperatorMessageList/OperatorMessageList";
import SupportMessageInput from "../SupportMessageInput/SupportMessageInput";

const TicketChat: React.FC = () => {
  const { store } = useContext(Context);
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();

  const ticketId = Number(id);
  const [loading, setLoading] = useState(true);
  const [info, setInfo] = useState<ITicketInfo | null>(null);
  const [msgs, setMsgs] = useState<IChatMessage[]>([]);
  const [users, setUsers] = useState<Array<{ id: number; name: string }>>([]);
  const [replyTo, setReplyTo] = useState<IChatMessage | null>(null);

  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!id || isNaN(ticketId)) { navigate("/dashboard/support"); return; }
    if (!store.isAuth) { navigate("/auth/login"); return; }

    (async () => {
      try {
        const i = await SupportService.getTicketInfo(ticketId);
        setInfo(i);
        setUsers([
          { id: i.user_id, name: i.user_name },
          ...(i.operator_id ? [{ id: i.operator_id, name: i.operator_name! }] : [])
        ]);

        const history = await SupportService.getTicketMessages(ticketId);
        setMsgs(history);


        SupportSocketService.connect(ticketId);
        SupportSocketService.onMessage(m => setMsgs(p => [...p, m]));
        SupportSocketService.onDelete(id =>
          setMsgs(p => p.map(m => m.id === id ? { ...m, deleted_at: new Date().toISOString() } : m))
        );
      } catch (e) {
        console.error("TicketChat init error:", e);
        navigate("/dashboard/support");
      } finally {
        setLoading(false);
      }
    })();

    return () => SupportSocketService.disconnect();
  }, [ticketId, store.isAuth]);

  useEffect(() => bottomRef.current?.scrollIntoView({ behavior: "smooth" }), [msgs]);

  const isOp = store.user?.role === "operator" || store.user?.role === "admin";

  const assign = async () => {
    await SupportService.assignTicket(ticketId);
    setInfo(p => p ? { ...p, status: "active", operator_id: store.user!.id, operator_name: store.user!.name } : p);
    setUsers(p => p.some(u => u.id === store.user!.id) ? p : [...p, { id: store.user!.id, name: store.user!.name }]);
  };

  const close = async () => {
    await SupportService.closeTicket(ticketId);
    setInfo(p => p ? { ...p, status: "closed" } : p);
  };

  const act = (a: "reply" | "delete", m: IChatMessage) => {
    if (a === "reply") setReplyTo(m);
    if (a === "delete") SupportSocketService.deleteMessage(ticketId, m.id);
  };

  if (loading || !info) return <div className={styles.loading}>Загрузка…</div>;

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h2 className={styles.subject}>{info.subject}</h2>
        <div className={styles.controls}>
          {isOp && info.status === "new" && (
            <button className={styles.assignBtn} onClick={assign}>Взять в работу</button>
          )}
          {(isOp || store.user!.id === info.user_id) && info.status === "active" && (
            <button className={styles.closeBtn} onClick={close}>Закрыть тикет</button>
          )}
          <span className={`${styles.status} ${styles[info.status]}`}>
            {info.status === "new" ? "Новый" : info.status === "active" ? "В работе" : "Закрыт"}
          </span>
        </div>
      </header>

      <main className={styles.messages}>
        {isOp ? (
          <OperatorMessageList
            messages={msgs}
            currentUserId={store.user!.id}
            operatorId={info.operator_id!}
            users={users}
            onAction={act}
          />
        ) : (
          <MessageList
            messages={msgs}
            currentUserId={store.user!.id}
            users={users}
            onAction={act}
          />
        )}
        <div ref={bottomRef} />
      </main>

      {info.status !== "closed" && (
        <SupportMessageInput
          replyTo={replyTo}
          onCancelReply={() => setReplyTo(null)}
          ticketId={ticketId}
        />
      )}
    </div>
  );
};

export default TicketChat;
