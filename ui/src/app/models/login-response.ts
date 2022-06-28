import { User } from "./user";

export interface LoginResponse {
    access_token: string;
    expired_at: string;
    user: User
  }
