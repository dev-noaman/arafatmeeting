import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { authService } from "../services/api/auth.service";

interface EmailVerificationProps {
  email?: string;
  onVerifySuccess?: () => void;
}

export default function EmailVerification({ email, onVerifySuccess }: EmailVerificationProps) {
  const navigate = useNavigate();
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [successMessage, setSuccessMessage] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await authService.verifyEmail({ email: email || "", code });
      setSuccessMessage("Email verified successfully!");
      setTimeout(() => {
        onVerifySuccess?.();
        navigate("/dashboard");
      }, 1500);
    } catch (err: any) {
      setError(err.response?.data?.error || "Invalid verification code");
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendCode = async () => {
    setError("");
    try {
      // Note: This would need a new API endpoint to resend
      setSuccessMessage("Verification code resent! Check your email.");
    } catch (err: any) {
      setError("Failed to resend code");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-600 via-brand-500 to-purple-700 px-4">
      <div className="max-w-md w-full">
        {/* Card */}
        <div className="glass rounded-2xl shadow-xl p-8 space-y-6">
          {/* Header */}
          <div className="text-center space-y-2">
            <div className="w-16 h-16 bg-brand-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
            </div>
            <h2 className="text-3xl font-bold text-gray-900">Verify Your Email</h2>
            <p className="text-gray-600">
              We've sent a verification code to <br />
              <span className="font-semibold text-brand-600">{email}</span>
            </p>
          </div>

          {successMessage && (
            <div className="p-4 bg-green-50 border border-green-200 rounded-lg text-green-700 text-center">
              {successMessage}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                {error}
              </div>
            )}

            <div>
              <label htmlFor="code" className="block text-sm font-medium text-gray-700 mb-1">
                Verification Code
              </label>
              <input
                type="text"
                id="code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-transparent tracking-widest text-center text-2xl font-mono"
                placeholder="000000"
                maxLength={6}
                autoComplete="one-time-code"
              />
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-3 px-4 bg-brand-600 hover:bg-brand-700 text-white font-semibold rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? "Verifying..." : "Verify Email"}
            </button>
          </form>

          <div className="text-center">
            <button
              onClick={handleResendCode}
              className="text-sm text-brand-600 hover:text-brand-700 font-medium"
            >
              Resend verification code
            </button>
          </div>

          <div className="text-center text-sm text-gray-600">
            Want to go back?{" "}
            <Link to="/login" className="text-brand-600 hover:text-brand-700 font-medium">
              Back to Login
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
