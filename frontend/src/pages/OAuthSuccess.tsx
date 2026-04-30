import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Loading } from "../components/common/Loading";
import { insforge } from "../services/insforge/client";
import type { UserSchema } from "@insforge/sdk";
import type { User } from "../types/user.types";

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

const OAuthSuccess = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(true);

  useEffect(() => {
    const processOAuth = async () => {
      try {
        const { data, error: authError } = await insforge.auth.getCurrentUser();
        if (authError || !data.user) {
          setError("No session found after OAuth redirect");
          setIsProcessing(false);
          setTimeout(() => navigate("/login"), 3000);
          return;
        }
        login(mapUser(data.user));
        setIsProcessing(false);
        navigate("/dashboard", { replace: true });
      } catch (err) {
        setError(err instanceof Error ? err.message : "OAuth failed");
        setIsProcessing(false);
        setTimeout(() => navigate("/login"), 3000);
      }
    };
    processOAuth();
  }, [navigate, login]);

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-linear-to-br from-brand-50 via-white to-accent-50">
        <div className="glass rounded-2xl shadow-xl p-8 max-w-md w-full text-center space-y-4">
          <div className="text-red-600 text-5xl">⚠️</div>
          <h2 className="text-2xl font-bold text-gray-900">
            Authentication Failed
          </h2>
          <p className="text-gray-600">{error}</p>
          <p className="text-sm text-gray-500">Redirecting to login...</p>
        </div>
      </div>
    );
  }

  if (isProcessing) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-linear-to-br from-brand-50 via-white to-accent-50">
        <div className="glass rounded-2xl shadow-xl p-8 max-w-md w-full text-center space-y-4">
          <Loading />
          <h2 className="text-2xl font-bold text-gray-900">
            Completing Sign In
          </h2>
          <p className="text-gray-600">
            Please wait while we set up your account...
          </p>
        </div>
      </div>
    );
  }

  return null;
};

export default OAuthSuccess;
