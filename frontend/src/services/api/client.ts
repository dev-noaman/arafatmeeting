import axios, { type AxiosError, type AxiosInstance } from "axios";
import API_BASE_URL from "../../utils/constants";
import { insforge } from "../insforge/client";

const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

apiClient.interceptors.request.use((config) => {
  const headers = insforge.getHttpClient().getHeaders();
  const auth = headers["Authorization"] || headers["authorization"];
  if (auth) {
    config.headers.Authorization = auth;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      const currentPath = window.location.pathname;
      const isAuthPage =
        currentPath === "/login" ||
        currentPath === "/register" ||
        currentPath === "/verify-email" ||
        currentPath === "/forgot-password" ||
        currentPath === "/reset-password" ||
        currentPath.startsWith("/auth/oauth");

      if (!isAuthPage) {
        console.warn("401 Unauthorized — redirecting to login");
        insforge.auth.signOut();
        window.location.href = "/login";
      }
    }

    if (!error.response) {
      console.error("Network error:", error.message);
    }

    return Promise.reject(error);
  },
);

export default apiClient;
