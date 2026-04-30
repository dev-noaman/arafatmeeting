import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { authService } from "../../services/api/auth.service";
import { useAuth } from "../../hooks/useAuth";

export const LoginForm: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  });
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleLocalLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      const { user } = await authService.login({
        email: formData.email,
        password: formData.password,
      });
      login(user);
      navigate("/dashboard");
    } catch (err: any) {
      setError(err.response?.data?.error || "Login failed");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md space-y-8">
      {/* Header */}
      <div className="text-center space-y-2">
        <h2 className="text-4xl font-bold text-gray-900 animate-slide-up">
          Welcome back
        </h2>
        <p
          className="text-lg text-gray-600 animate-slide-up"
          style={{ animationDelay: "100ms" }}
        >
          Sign in to your account
        </p>
      </div>

      {/* Glassmorphic Form Card */}
      <div
        className="glass rounded-2xl shadow-xl p-8 space-y-6 animate-scale-up"
        style={{ animationDelay: "200ms" }}
      >
        {/* Local Login Form */}
        <form onSubmit={handleLocalLogin} className="space-y-4">
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
              {error}
            </div>
          )}

          <div>
            <label
              htmlFor="email"
              className="block text-sm font-medium text-gray-700 mb-1"
            >
              Email
            </label>
            <input
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleInputChange}
              required
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-transparent"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label
              htmlFor="password"
              className="block text-sm font-medium text-gray-700 mb-1"
            >
              Password
            </label>
            <div className="flex justify-between items-center">
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleInputChange}
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-transparent"
                placeholder="Enter your password"
              />
            </div>
            <div className="flex justify-end mt-1">
              <Link
                to="/forgot-password"
                className="text-sm text-brand-600 hover:text-brand-700 font-medium"
              >
                Forgot password?
              </Link>
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full py-3 px-4 bg-brand-600 hover:bg-brand-700 text-white font-semibold rounded-lg transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? "Signing in..." : "Sign In"}
          </button>
        </form>
      </div>

      {/* Sign Up Link */}
      <div className="text-center">
        <span className="text-gray-600">Don't have an account? </span>
        <Link
          to="/register"
          className="font-semibold text-brand-600 hover:text-brand-700 transition-colors"
        >
          Sign up for free
        </Link>
      </div>
    </div>
  );
};
