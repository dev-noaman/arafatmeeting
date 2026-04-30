import { useAuthStore } from "../store";

export const useAuth = () => {
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoading = useAuthStore((state) => state.isLoading);
  const logout = useAuthStore((state) => state.logout);
  const login = useAuthStore((state) => state.login);
  const setUser = useAuthStore((state) => state.setUser);

  return { user, isAuthenticated, isLoading, logout, login, setUser };
};
