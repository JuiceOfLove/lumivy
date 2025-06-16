export interface IChatMessage {
  id:          number;
  user_id:     number;
  sender_id?:  number;
  content?:    string;
  media_url?:  string;
  reply_to_id?: number | null;
  created_at:  string;
  deleted_at?: string | null;
}