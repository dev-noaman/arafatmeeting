import { describe, it, expect, beforeEach, vi } from "vitest";
import { createAuthActions } from "./authActions";

vi.mock("../../services/insforge/client", () => ({
  insforge: {
    auth: {
      signOut: vi.fn().mockResolvedValue(undefined),
      getCurrentUser: vi.fn(),
    },
  },
}));

describe("createAuthActions", () => {
  let setState: Partial<{ user: unknown; isAuthenticated: boolean; isLoading: boolean }>;
  let set: ReturnType<typeof vi.fn>;
  let get: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    setState = {};
    set = vi.fn((partial) => {
      setState = { ...setState, ...partial };
    });
    get = vi.fn(() => ({
      user: (setState as any).user ?? null,
      isAuthenticated: (setState as any).isAuthenticated ?? false,
      isLoading: (setState as any).isLoading ?? true,
    }));
  });

  it("setUser sets user and isAuthenticated", () => {
    const actions = createAuthActions(set, get);
    const user = { id: "u1", email: "test@test.com", name: "Test", role: "user" };
    actions.setUser(user as any);
    expect(set).toHaveBeenCalledWith({ user, isAuthenticated: true });
  });

  it("setUser with null clears auth", () => {
    const actions = createAuthActions(set, get);
    actions.setUser(null);
    expect(set).toHaveBeenCalledWith({ user: null, isAuthenticated: false });
  });

  it("login sets user, isAuthenticated, and loading false", () => {
    const actions = createAuthActions(set, get);
    const user = { id: "u1", email: "t@t.com", name: "T", role: "user" };
    actions.login(user as any);
    expect(set).toHaveBeenCalledWith({ user, isAuthenticated: true, isLoading: false });
  });
});
