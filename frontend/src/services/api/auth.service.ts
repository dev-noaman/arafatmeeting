import { insforge } from "../insforge/client";
import type { User } from "../../types/user.types";
import type { UserSchema } from "@insforge/sdk";

function mapUser(u: UserSchema | undefined | null): User | null {
  if (!u) return null;
  return {
    id: u.id,
    email: u.email,
    name: u.profile?.name || "",
    role: "user",
    provider: (u.providers?.[0] as User["provider"]) || "local",
    created_at: u.createdAt,
  };
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface VerifyEmailRequest {
  email: string;
  code: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  email: string;
  code: string;
  password: string;
}

export const authService = {
  login: async (request: LoginRequest): Promise<{ token: string; user: User }> => {
    const { data, error } = await insforge.auth.signInWithPassword({
      email: request.email,
      password: request.password,
    });
    if (error) throw new Error(error.message || "Login failed");
    if (!data?.user) throw new Error("Login failed: no user data returned");
    const user = mapUser(data.user);
    if (!user) throw new Error("Login failed: invalid user data");
    return {
      token: data.accessToken,
      user,
    };
  },

  register: async (request: RegisterRequest): Promise<{ message: string; user: { id: string; email: string; name: string } }> => {
    const { data, error } = await insforge.auth.signUp({
      email: request.email,
      password: request.password,
      name: request.name,
    });
    if (error) throw new Error(error.message || "Registration failed");
    // InsForge signUp returns {accessToken, requireEmailVerification} — user object may be absent
    const userId = data?.user?.id ?? "pending";
    const userEmail = data?.user?.email ?? request.email;
    return {
      message: "User registered successfully",
      user: { id: userId, email: userEmail, name: request.name },
    };
  },

  initiateOAuthLogin: (provider: "google" | "github"): void => {
    insforge.auth.signInWithOAuth({
      provider,
      redirectTo: `${window.location.origin}/auth/oauth-success`,
    });
  },

  verifyEmail: async (request: VerifyEmailRequest): Promise<{ message: string }> => {
    const { error } = await insforge.auth.verifyEmail({
      email: request.email,
      otp: request.code,
    });
    if (error) throw new Error(error.message || "Verification failed");
    return { message: "Email verified successfully" };
  },

  forgotPassword: async (request: ForgotPasswordRequest): Promise<{ message: string }> => {
    const { error } = await insforge.auth.sendResetPasswordEmail({
      email: request.email,
    });
    if (error) throw new Error(error.message || "Failed to send reset email");
    return { message: "Reset email sent" };
  },

  resetPassword: async (request: ResetPasswordRequest): Promise<{ message: string }> => {
    // Step 1: exchange the 6-digit code for a reset token
    const { data: tokenData, error: exchangeError } = await insforge.auth.exchangeResetPasswordToken({
      email: request.email,
      code: request.code,
    });
    if (exchangeError) throw new Error(exchangeError.message || "Invalid reset code");
    // Step 2: reset password with the token
    const { error } = await insforge.auth.resetPassword({
      newPassword: request.password,
      otp: tokenData!.token,
    });
    if (error) throw new Error(error.message || "Password reset failed");
    return { message: "Password reset successfully" };
  },
};
