import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { authService } from "../../services/api/auth.service";
import { useAuth } from "../../hooks/useAuth";

export const RegisterForm: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [formData, setFormData] = useState({
    email: "",
    password: "",
    confirmPassword: "",
    name: "",
  });
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (formData.password !== formData.confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    if (formData.password.length < 6) {
      setError("Password must be at least 6 characters");
      return;
    }

    setIsLoading(true);

    try {
      await authService.register({
        email: formData.email,
        password: formData.password,
        name: formData.name,
      });
      // Automatically log in after registration
      const { user } = await authService.login({
        email: formData.email,
        password: formData.password,
      });
      login(user);
      navigate("/dashboard");
    } catch (err: any) {
      console.error("[Register] Registration error:", err);
      console.error("[Register] Error response:", err.response);
      console.error("[Register] Error message:", err.message);
      const insforgeError = err.message || "Registration failed";
      setError(insforgeError);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md space-y-8">
      {/* Header */}
      <div className="text-center space-y-2">
        <h2 className="text-4xl font-bold text-gray-900 animate-slide-up">
          Create account
        </h2>
        <p className="text-lg text-gray-600 animate-slide-up" style={{ animationDelay: '100ms' }}>
          Join Mini Meeting today
        </p>
      </div>

      {/* Glassmorphic Form Card */}
      <div className="glass rounded-2xl shadow-xl p-8 space-y-6 animate-scale-up" style={{ animationDelay: '200ms' }}>
        {/* Registration Form */}
        <form onSubmit={handleRegister} className="space-y-4">
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
              {error}
            </div>
          )}

          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              Full Name
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              required
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-transparent"
              placeholder="John Doe"
            />
          </div>

          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
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
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              Password
            </label>
            <input
              type="password"
              id="password"
              name="password"
              value={formData.password}
              onChange={handleInputChange}
              required
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-transparent"
              placeholder="At least 6 characters"
            />
          </div>

          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
              Confirm Password
            </label>
            <input
              type="password"
              id="confirmPassword"
              name="confirmPassword"
              value={formData.confirmPassword}
              onChange={handleInputChange}
              required
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-transparent"
              placeholder="Re-enter your password"
            />
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full py-3 px-4 bg-brand-600 hover:bg-brand-700 text-white font-semibold rounded-lg transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? "Creating account..." : "Create Account"}
          </button>
        </form>
      </div>

      {/* Sign In Link */}
      <div className="text-center">
        <span className="text-gray-600">Already have an account? </span>
        <Link
          to="/login"
          className="font-semibold text-brand-600 hover:text-brand-700 transition-colors"
        >
          Sign in
        </Link>
      </div>
    </div>
  );
};
