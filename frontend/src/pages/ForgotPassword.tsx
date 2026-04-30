import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { authService } from "../services/api/auth.service";
import { Button } from "../components/common/Button";

export default function ForgotPassword() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await authService.forgotPassword({ email });
      setSuccess(true);
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to send reset email");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex">
      {/* Back to Home Button */}
      <div className="absolute top-4 left-4 z-20">
        <Button
          variant="ghost"
          onClick={() => navigate("/")}
          size="md"
          leftIcon={
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          }
        >
          Back to Home
        </Button>
      </div>

      {/* Left Side - Form */}
      <div className="flex-1 flex items-center justify-center px-4 sm:px-6 lg:px-8 py-12">
        <div className="w-full max-w-md space-y-8">
          <div className="text-center space-y-2">
            <h2 className="text-4xl font-bold text-gray-900 animate-slide-up">
              Forgot password?
            </h2>
            <p className="text-lg text-gray-600 animate-slide-up" style={{ animationDelay: "100ms" }}>
              Enter your email and we'll send a reset code.
            </p>
          </div>

          <div className="glass rounded-2xl shadow-xl p-8 space-y-6 animate-scale-up" style={{ animationDelay: "200ms" }}>
            {success ? (
              <div className="space-y-4">
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg text-green-700 text-center">
                  <svg className="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  <p className="font-semibold">Check your email!</p>
                  <p className="text-sm mt-1">
                    We've sent a password reset code to <strong>{email}</strong>
                  </p>
                </div>
                <Link
                  to="/login"
                  className="block w-full py-3 px-4 bg-brand-600 hover:bg-brand-700 text-white font-semibold rounded-lg text-center transition-colors duration-200"
                >
                  Back to Login
                </Link>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4">
                {error && (
                  <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                    {error}
                  </div>
                )}
                <div>
                  <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                    Email Address
                  </label>
                  <input
                    type="email"
                    id="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-transparent"
                    placeholder="you@example.com"
                  />
                </div>
                <button
                  type="submit"
                  disabled={isLoading}
                  className="w-full py-3 px-4 bg-brand-600 hover:bg-brand-700 text-white font-semibold rounded-lg transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isLoading ? "Sending..." : "Send Reset Code"}
                </button>
              </form>
            )}
          </div>

          <div className="text-center">
            <span className="text-gray-600">Remember your password? </span>
            <Link
              to="/login"
              className="font-semibold text-brand-600 hover:text-brand-700 transition-colors"
            >
              Back to Login
            </Link>
          </div>
        </div>
      </div>

      {/* Right Side - Visual/Branding */}
      <div className="hidden lg:flex lg:flex-1 relative overflow-hidden">
        <div className="absolute inset-0 bg-linear-to-br from-brand-500 via-purple-600 to-brand-700">
          <div className="absolute top-20 left-20 w-72 h-72 bg-white/10 rounded-full blur-3xl animate-pulse-slow" />
          <div className="absolute bottom-20 right-20 w-96 h-96 bg-white/10 rounded-full blur-3xl animate-pulse-slow" style={{ animationDelay: '1s' }} />
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-64 h-64 bg-white/5 rounded-full blur-2xl animate-bounce-gentle" />
        </div>

        <div className="relative z-10 flex flex-col items-center justify-center text-white p-12 w-full">
          <div className="mb-8 w-20 h-20 bg-white/20 backdrop-blur-lg rounded-2xl flex items-center justify-center shadow-2xl">
            <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>

          <h1 className="text-4xl font-bold text-center mb-4">
            Secure access,
            <br />
            always
          </h1>
          <p className="text-lg text-white/80 text-center max-w-md mb-12">
            Your account security is important. We'll help you get back in safely.
          </p>

          <div className="space-y-4 w-full max-w-md">
            {[
              { icon: '🔒', text: 'Encrypted password reset' },
              { icon: '📧', text: 'Code sent to your email' },
              { icon: '⚡', text: 'Quick and easy recovery' },
            ].map((feature, i) => (
              <div
                key={i}
                className="flex items-center gap-3 bg-white/10 backdrop-blur-sm rounded-lg p-4 animate-slide-in-left"
                style={{ animationDelay: `${i * 100}ms` }}
              >
                <div className="w-8 h-8 bg-white/20 rounded-full flex items-center justify-center text-lg">
                  {feature.icon}
                </div>
                <span className="text-white/90">{feature.text}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
