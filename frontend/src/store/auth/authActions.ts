import type { AuthActions } from "./auth.types";
import type { User } from "../../types/user.types";
import { insforge } from "../../services/insforge/client";
import type { UserSchema } from "@insforge/sdk";

function mapUser(u: UserSchema): User {
  return {
    id: u.id,
    email: u.email,
    name: u.profile?.name || "",
    role: "user",
    provider: (u.providers?.[0] as User["provider"]) || "local",
    created_at: u.createdAt,
  };
}

export const createAuthActions = (
  set: (partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean }>) => void,
  get: () => { user: User | null; isAuthenticated: boolean; isLoading: boolean },
): AuthActions => ({
  setUser: (user: User | null) => {
    set({ user, isAuthenticated: !!user });
  },

  login: (user: User) => {
    set({ user, isAuthenticated: true, isLoading: false });
  },

  logout: async () => {
    await insforge.auth.signOut();
    set({ user: null, isAuthenticated: false });
  },

  initialize: async () => {
    try {
      console.debug("[AuthInit] Calling getCurrentUser()...");
      const { data, error } = await insforge.auth.getCurrentUser();
      console.debug("[AuthInit] getCurrentUser result:", {
        hasUser: !!data?.user,
        error: error ? { message: error.message, statusCode: error.statusCode, error: error.error } : null,
      });
      if (error || !data.user) {
        set({ user: null, isAuthenticated: false, isLoading: false });
        return;
      }
      set({ user: mapUser(data.user), isAuthenticated: true, isLoading: false });
    } catch (err) {
      console.error("[AuthInit] Unexpected error in initialize:", err);
      set({ user: null, isAuthenticated: false, isLoading: false });
    }
  },
});
