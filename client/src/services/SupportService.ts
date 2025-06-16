import $api from "../http";
import { IMyTicket, IOperatorTicket, ITicketInfo } from "../types/support";
import { IChatMessage } from "../types/chat";

class SupportService {
  static async createTicket(payload: { subject: string; content: string })
    : Promise<{ ticket_id: number }> {
    const { data } = await $api.post("/support/tickets", payload);
    return data;
  }

  static async getMyTickets(): Promise<IMyTicket[]> {
    const { data } = await $api.get("/support/tickets/my");
    return data;
  }

  static async getTicketInfo(ticketId: number): Promise<ITicketInfo> {
    const { data } = await $api.get(`/support/tickets/${ticketId}`);
    return data;
  }

  static async getTicketMessages(ticketId: number): Promise<IChatMessage[]> {
    const { data } = await $api.get(`/support/tickets/${ticketId}/messages`);
    return data.map((m: any) => ({ ...m, user_id: m.sender_id }));
  }


  static async assignTicket(id: number) { await $api.post(`/support/tickets/${id}/assign`); }
  static async closeTicket(id: number) { await $api.post(`/support/tickets/${id}/close`); }

  static async getOperatorTickets(status: "new" | "active" | "closed")
    : Promise<IOperatorTicket[]> {
    const { data } = await $api.get("/support/tickets/operator/list",
      { params: { status } });
    return data;
  }

  static async deleteTicketMessage(ticketId: number, msgId: number) {
    await $api.delete(`/support/tickets/${ticketId}/messages/${msgId}`);
  }
}

export default SupportService;
