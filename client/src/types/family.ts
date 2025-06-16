import { IUser } from "./auth";

export interface IFamily {
  id: number;
  name: string;
  owner_id: number;
  created_at: string;
  updated_at: string;
}

export interface IFamilyResponse {
  family: IFamily;
  members: IUser[];
}

export interface IInviteResponse {
  message: string;
}