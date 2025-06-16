import $api from "../http";
import { IChatMessage } from "../types/chat";

export type MsgHandler = (msg: IChatMessage) => void;
export type DelHandler = (id: number) => void;

class SupportSocketService {
  private ws: WebSocket | null = null;

  private msgHandlers: MsgHandler[] = [];
  private delHandlers: DelHandler[] = [];

  onMessage(cb: MsgHandler) { this.msgHandlers.push(cb); }
  onDelete(cb: DelHandler) { this.delHandlers.push(cb); }

  connect(ticketId: number) {
    if (this.ws) this.ws.close();

    const token = localStorage.getItem("token");
    if (!token) return;

    const proto = location.protocol === "https:" ? "wss" : "ws";
    const host = location.port === "5173" ? "localhost:8080" : location.host;
    const url = `${proto}://${host}/api/support/ws/${ticketId}?token=${token}`;

    this.ws = new WebSocket(url);

    this.ws.addEventListener("open", () =>
      console.log("%cWS(ticket) OPEN", "color:green", url)
    );

    this.ws.addEventListener("error", () => console.error("WS(ticket) ERROR"));
    this.ws.addEventListener("close", () => console.warn("WS(ticket) CLOSED"));

    this.ws.addEventListener("message", e => {
      const { event, data } = JSON.parse(e.data);

      if (event === "support:ticket_message") {
        const msg = { ...data, user_id: data.sender_id } as IChatMessage;
        this.msgHandlers.forEach(cb => cb(msg));
      }

      if (event === "support:ticket_delete") {
        this.delHandlers.forEach(cb => cb(data.message_id));
      }
    });
  }

  disconnect() {
    this.ws?.close();
    this.ws = null;
    this.msgHandlers = [];
    this.delHandlers = [];
  }

  async send(text: string, ticketId: number,
    replyTo?: number, mediaB64?: string) {

    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const payload: any = { content: text };
      if (replyTo) payload.reply_to = replyTo;
      if (mediaB64) payload.media = mediaB64;
      try { this.ws.send(JSON.stringify(payload)); return; }
      catch (e) { console.warn("WS send error, fallback → HTTP"); }
    }

    const body: any = { content: text };
    if (replyTo) body.reply_to_id = replyTo;
    if (mediaB64) body.media_url = mediaB64;

    await $api.post(`/support/tickets/${ticketId}/messages`, body);
  }

  async deleteMessage(ticketId: number, messageId: number) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify({ delete_id: messageId }));
        return;
      } catch (e) { }
    }
    await $api.delete(`/support/tickets/${ticketId}/messages/${messageId}`);
  }
}

export default new SupportSocketService();
