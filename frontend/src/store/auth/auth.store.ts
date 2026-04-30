import { create } from "zustand";
import { devtools } from "zustand/middleware";
import type { AuthStore } from "./auth.types";
import { createAuthActions } from "./authActions";

export const useAuthStore = create<AuthStore>()(
  devtools((set, get) => ({
    user: null,
    isAuthenticated: false,
    isLoading: true,
    ...createAuthActions(set, get),
  })),
);
