import { insforge } from "../../insforge/client";
import type { User, PaginatedUsersResponse } from "../../../types/user.types";
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

export const userAdminService = {
  getAllUsers: async (page: number = 1, pageSize: number = 10, search: string = ""): Promise<PaginatedUsersResponse> => {
    const offset = (page - 1) * pageSize;
    const { data, error, count } = await insforge.database
      .from("users")
      .select("*", { count: "exact" })
      .range(offset, offset + pageSize - 1)
      .order("created_at", { ascending: false });
    if (error) throw new Error(error.message);
    const users = ((data || []) as UserSchema[]).map(mapUser);
    const filtered = search
      ? users.filter((u) => u.email.includes(search) || u.name.includes(search))
      : users;
    return {
      data: filtered,
      total: count || 0,
      page,
      page_size: pageSize,
      total_pages: Math.ceil((count || 0) / pageSize),
    };
  },

  getUserById: async (id: string): Promise<User> => {
    const { data, error } = await insforge.auth.getProfile(id);
    if (error) throw new Error(error.message);
    const u = data! as UserSchema;
    return {
      id: u.id,
      email: u.email || "",
      name: u.profile?.name || "",
      role: "user",
      provider: (u.providers?.[0] as User["provider"]) || "local",
      created_at: u.createdAt || "",
    };
  },

  deleteUser: async (id: string): Promise<void> => {
    // InsForge doesn't expose admin user deletion via SDK directly
    // Use the HTTP client to call the admin API
    await insforge.getHttpClient().delete(`/api/auth/admin/users/${id}`);
  },
};
