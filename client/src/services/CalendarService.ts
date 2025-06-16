import $api from "../http";
import { IEvent } from "../types/EventTypes";

export default class CalendarService {
  static async getAllCalendars() {
    const response = await $api.get("/calendar/list");
    return response.data;
  }

  static async createExtraCalendar(title: string) {
    const response = await $api.post("/calendar/create_extra", { title });
    return response.data.calendar;
  }

  static async getTodayEvents(): Promise<IEvent[]> {
    const res = await $api.get<IEvent[]>("/calendar/today");
    return res.data;
  }

  static async getNextEvent(): Promise<IEvent> {
    const res = await $api.get<IEvent>("/calendar/next");
    return res.data;
  }
}