import { useState } from "react";
import { Button } from "../../common/Button";
import { Input } from "../../common/Input";

interface AttendeeInputProps {
  attendees: string[];
  onAddAttendee: (email: string) => void;
  onRemoveAttendee: (email: string) => void;
}

export const AttendeeInput: React.FC<AttendeeInputProps> = ({
  attendees,
  onAddAttendee,
  onRemoveAttendee,
}) => {
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleAdd = () => {
    if (!email.trim()) {
      setError("Email is required");
      return;
    }

    if (!validateEmail(email.trim())) {
      setError("Please enter a valid email address");
      return;
    }

    if (attendees.includes(email.trim())) {
      setError("This email is already added");
      return;
    }

    onAddAttendee(email.trim());
    setEmail("");
    setError("");
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAdd();
    }
  };

  return (
    <div className="mb-6">
      <label className="block text-sm font-medium text-gray-700 mb-2">
        Invite Attendees (optional)
      </label>
      <div className="flex gap-2 mb-3">
        <div className="flex-1">
          <Input
            type="email"
            value={email}
            onChange={(e) => {
              setEmail(e.target.value);
              setError("");
            }}
            onKeyPress={handleKeyPress}
            placeholder="colleague@example.com"
            error={error}
            leftIcon={
              <svg
                className="h-5 w-5 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                />
              </svg>
            }
          />
        </div>
        <Button type="button" onClick={handleAdd} variant="secondary" size="md">
          <svg
            className="h-5 w-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          Add
        </Button>
      </div>
      {error && <p className="text-sm text-red-600 mb-2">{error}</p>}

      {attendees.length > 0 && (
        <div className="mt-3 space-y-2">
          <p className="text-sm text-gray-500">
            {attendees.length} attendee{attendees.length > 1 ? "s" : ""} invited:
          </p>
          <div className="flex flex-wrap gap-2">
            {attendees.map((attendeeEmail) => (
              <span
                key={attendeeEmail}
                className="inline-flex items-center gap-1 px-3 py-1 bg-blue-50 text-blue-700 rounded-full text-sm"
              >
                {attendeeEmail}
                <button
                  type="button"
                  onClick={() => onRemoveAttendee(attendeeEmail)}
                  className="hover:text-blue-900 transition-colors"
                  aria-label={`Remove ${attendeeEmail}`}
                >
                  <svg
                    className="h-4 w-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
