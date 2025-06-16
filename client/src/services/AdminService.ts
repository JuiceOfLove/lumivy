import $api from "../http";
import { IUser } from "../types/auth";

export default class AdminService {
  static async getPaymentHistory() {
    const { data } = await $api.get("/admin/payments");
    return data;
  }

  static async listOperators(): Promise<IUser[]> {
    const { data } = await $api.get("/admin/operators");
    return data;
  }

  static async searchUsers(q: string): Promise<IUser[]> {
    const { data } = await $api.get("/admin/users/search", { params: { q } });
    return data;
  }

  static async makeOperator(id: number) {
    await $api.post(`/admin/users/${id}/make_operator`);
  }
}