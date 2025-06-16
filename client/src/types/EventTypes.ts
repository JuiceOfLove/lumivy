export interface IEvent {
    id: number;
    family_id: number;
    title: string;
    description: string;
    start_time: string;
    end_time: string;
    created_by: number;
    is_completed: boolean;
    color?: string | null;
    private: boolean;
    created_at: string;
    updated_at: string;
  }

  export interface ICreateEventRequest {
    calendar_id: number;
    title: string;
    description: string;
    start_time: string;
    end_time: string;
    color?: string;
    private?: boolean;
  }