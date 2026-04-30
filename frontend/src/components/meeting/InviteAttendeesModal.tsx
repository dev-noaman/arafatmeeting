import { useState } from "react";
import { Modal } from "../common/Modal";
import { Button } from "../common/Button";
import { Input } from "../common/Input";
import { attendeeService } from "../../services/api/attendee.service";
import type { AttendeeResponse } from "../../types/meeting.types";

interface InviteAttendeesModalProps {
  isOpen: boolean;
  onClose: () => void;
  meetingId: number;
  existingAttendees?: AttendeeResponse[];
}

export const InviteAttendeesModal: React.FC<InviteAttendeesModalProps> = ({
  isOpen,
  onClose,
  meetingId,
  existingAttendees = [],
}) => {
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [attendees, setAttendees] = useState<AttendeeResponse[]>(
    existingAttendees
  );
  const [isSaving, setIsSaving] = useState(false);

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleAdd = async () => {
    if (!email.trim()) {
      setError("Email is required");
      return;
    }

    if (!validateEmail(email.trim())) {
      setError("Please enter a valid email address");
      return;
    }

    if (attendees.some((a) => a.email === email.trim())) {
      setError("This email is already added");
      return;
    }

    setIsSaving(true);
    try {
      const newAttendee = await attendeeService.addAttendee(
        meetingId,
        email.trim()
      );
      setAttendees((prev) => [...prev, newAttendee]);
      setEmail("");
      setError("");
    } catch (err) {
      console.error("Failed to add attendee:", err);
      setError("Failed to add attendee");
    } finally {
      setIsSaving(false);
    }
  };

  const handleRemove = async (attendee: AttendeeResponse) => {
    setIsSaving(true);
    try {
      await attendeeService.removeAttendee(meetingId, attendee.email);
      setAttendees((prev) => prev.filter((a) => a.id !== attendee.id));
    } catch (err) {
      console.error("Failed to remove attendee:", err);
      setError("Failed to remove attendee");
    } finally {
      setIsSaving(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAdd();
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="md" showCloseButton={true}>
      <div className="p-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-4">
          Invite Attendees
        </h2>
        <p className="text-gray-600 mb-6">
          Add or remove attendees for this meeting.
        </p>

        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Add Attendee by Email
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
            <Button
              type="button"
              onClick={handleAdd}
              variant="secondary"
              size="md"
              disabled={isSaving}
            >
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
                {attendees.length} attendee{attendees.length > 1 ? "s" : ""}:
              </p>
              <div className="flex flex-wrap gap-2">
                {attendees.map((attendee) => (
                  <span
                    key={attendee.id}
                    className="inline-flex items-center gap-1 px-3 py-1 bg-blue-50 text-blue-700 rounded-full text-sm"
                  >
                    {attendee.email}
                    <button
                      type="button"
                      onClick={() => handleRemove(attendee)}
                      className="hover:text-blue-900 transition-colors disabled:opacity-50"
                      disabled={isSaving}
                      aria-label={`Remove ${attendee.email}`}
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

        <div className="flex gap-3">
          <Button onClick={onClose} variant="primary" fullWidth>
            Done
          </Button>
        </div>
      </div>
    </Modal>
  );
};
