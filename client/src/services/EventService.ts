import $api from "../http";
import { IEvent, ICreateEventRequest } from "../types/EventTypes";

export default class EventService {
  static async createEvent(data: ICreateEventRequest): Promise<IEvent> {
    const res = await $api.post<{ event: IEvent }>("/calendar/events", data);
    return res.data.event;
  }

  static async getAllEvents(): Promise<IEvent[]> {
    const res = await $api.get<IEvent[]>("/calendar/events/all");
    return res.data;
  }

  static async getEventsForMonth(month: number, year: number): Promise<IEvent[]> {
    const res = await $api.get<IEvent[]>("/calendar/events", { params: { month, year } });
    return res.data;
  }

  static async completeEvent(id: number): Promise<IEvent> {
    const res = await $api.post<{ event: IEvent }>(`/calendar/events/${id}/complete`);
    return res.data.event;
  }

  static async updateEvent(id: number, data: ICreateEventRequest): Promise<IEvent> {
    const res = await $api.put<{ event: IEvent }>(`/calendar/events/${id}`, data);
    return res.data.event;
  }

  static async getEventsForCalendar(calendarId: number, month: number, year: number): Promise<IEvent[]> {
    const res = await $api.get<IEvent[]>(`/calendar/${calendarId}/events`, {
      params: { month, year },
    });
    return res.data;
  }
}