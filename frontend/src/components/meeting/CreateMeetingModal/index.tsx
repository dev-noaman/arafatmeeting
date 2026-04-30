import { useState } from "react";
import { Modal } from "../../common/Modal";
import { Button } from "../../common/Button";
import { ErrorMessage } from "../../common/ErrorMessage";
import { MeetingInfoDisplay } from "./MeetingInfoDisplay";
import { AttendeeInput } from "./AttendeeInput";
import { useCreateMeeting } from "./useCreateMeeting";
import type { AttendeeResponse } from "../../../types/meeting.types";

interface CreateMeetingModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const CreateMeetingModal: React.FC<CreateMeetingModalProps> = ({
  isOpen,
  onClose,
}) => {
  const [attendees, setAttendees] = useState<string[]>([]);
  const [isCreating, setIsCreating] = useState(false);

  const {
    meeting,
    isLoading,
    error,
    copied,
    handleCreateMeeting,
    handleCopyLink,
    reset,
  } = useCreateMeeting({ attendees });

  const handleClose = () => {
    reset();
    setAttendees([]);
    onClose();
  };

  const handleAddAttendee = (email: string) => {
    setAttendees((prev) => [...prev, email]);
  };

  const handleRemoveAttendee = (email: string) => {
    setAttendees((prev) => prev.filter((e) => e !== email));
  };

  const handleCreateWithAttendees = async () => {
    setIsCreating(true);
    await handleCreateMeeting();
    setIsCreating(false);
  };

  // Show attendee input on first open, then show meeting info after creation
  const showAttendeeInput = !meeting && !isLoading && !error;

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      size="md"
      showCloseButton={true}
    >
      <div className="p-6">
        {showAttendeeInput ? (
          <div>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">
              Create New Meeting
            </h2>
            <p className="text-gray-600 mb-6">
              Create a meeting and optionally invite attendees by email.
            </p>

            <AttendeeInput
              attendees={attendees}
              onAddAttendee={handleAddAttendee}
              onRemoveAttendee={handleRemoveAttendee}
            />

            <div className="flex gap-3 mt-6">
              <Button
                onClick={handleCreateWithAttendees}
                variant="primary"
                fullWidth
                disabled={isCreating}
              >
                {isCreating ? "Creating..." : "Create Meeting"}
              </Button>
              <Button onClick={handleClose} variant="secondary" fullWidth>
                Cancel
              </Button>
            </div>
          </div>
        ) : isLoading ? (
          <div className="text-center py-8">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <p className="mt-4 text-gray-600">Creating your meeting...</p>
          </div>
        ) : error ? (
          <div>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Error</h2>
            <ErrorMessage message={error} />
            <div className="mt-6 flex gap-3">
              <Button onClick={handleCreateWithAttendees} variant="primary" fullWidth>
                Try Again
              </Button>
              <Button onClick={handleClose} variant="secondary" fullWidth>
                Cancel
              </Button>
            </div>
          </div>
        ) : meeting ? (
          <MeetingInfoDisplay
            meetingLink={meeting.meeting_link}
            meetingCode={meeting.meeting_code}
            copied={copied}
            onCopy={handleCopyLink}
            onClose={handleClose}
            attendees={meeting.attendees as AttendeeResponse[] | undefined}
          />
        ) : null}
      </div>
    </Modal>
  );
};
