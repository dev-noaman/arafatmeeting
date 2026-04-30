import type { User } from "../../types/user.types";

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface AuthActions {
  setUser: (user: User | null) => void;
  login: (user: User) => void;
  logout: () => Promise<void>;
  initialize: () => Promise<void>;
}

export type AuthStore = AuthState & AuthActions;
