import { insforge } from "../../insforge/client";
import type { User, UpdateUserRequest } from "../../../types/user.types";
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

export const userProfileService = {
  getCurrentUser: async (): Promise<User> => {
    const { data, error } = await insforge.auth.getCurrentUser();
    if (error || !data.user) throw new Error("Not authenticated");
    return mapUser(data.user);
  },

  updateCurrentUser: async (updateData: UpdateUserRequest): Promise<User> => {
    const { data, error } = await insforge.auth.setProfile({ name: updateData.name });
    if (error) throw new Error(error.message);
    // Re-fetch user to get updated state
    const { data: freshData, error: freshError } = await insforge.auth.getCurrentUser();
    if (freshError || !freshData.user) throw new Error("Failed to refresh user");
    return mapUser(freshData.user);
  },
};
