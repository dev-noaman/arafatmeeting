import type { AttendeeResponse } from "../../../types/meeting.types";
import { Button } from "../../common/Button";

interface MeetingInfoDisplayProps {
  meetingLink: string;
  meetingCode: string;
  copied: boolean;
  onCopy: () => void;
  onClose: () => void;
  attendees?: AttendeeResponse[];
}

export const MeetingInfoDisplay: React.FC<MeetingInfoDisplayProps> = ({
  meetingLink,
  meetingCode,
  copied,
  onCopy,
  onClose,
  attendees = [],
}) => (
  <div>
    <h2 className="text-2xl font-bold text-gray-900 mb-2">
      Here's your joining info
    </h2>
    <p className="text-gray-600 mb-6">
      Send this to people you want to meet with. Be sure to save it so you can
      use it later, too.
    </p>
    <div className="bg-gray-50 rounded-lg p-4 mb-6">
      <div className="flex items-center justify-between">
        <div className="flex-1 mr-3">
          <p className="text-sm text-gray-500 mb-1">Meeting link</p>
          <p className="text-blue-600 font-medium break-all">{meetingLink}</p>
        </div>
        <button
          onClick={onCopy}
          className="shrink-0 p-2 hover:bg-gray-200 rounded-lg transition-colors"
          title="Copy link"
        >
          {copied ? (
            <svg
              className="h-6 w-6 text-green-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          ) : (
            <svg
              className="h-6 w-6 text-gray-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
              />
            </svg>
          )}
        </button>
      </div>
      <div className="mt-3 pt-3 border-t border-gray-200">
        <p className="text-sm text-gray-500 mb-1">Meeting code</p>
        <p className="text-gray-900 font-mono font-semibold text-lg">
          {meetingCode}
        </p>
      </div>
    </div>

    {attendees.length > 0 && (
      <div className="mb-6">
        <h3 className="text-sm font-medium text-gray-700 mb-2">
          Invited Attendees
        </h3>
        <div className="bg-green-50 border border-green-200 rounded-lg p-3">
          <div className="flex items-start gap-2">
            <svg
              className="h-5 w-5 text-green-600 mt-0.5"
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
            <div className="flex-1">
              <p className="text-sm text-green-800">
                Invitation emails will be sent to:
              </p>
              <ul className="mt-2 space-y-1">
                {attendees.map((attendee) => (
                  <li
                    key={attendee.id}
                    className="text-sm text-green-900 flex items-center gap-2"
                  >
                    <span className="w-1.5 h-1.5 bg-green-600 rounded-full"></span>
                    {attendee.email}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </div>
    )}

    <div className="flex gap-3">
      <Button onClick={onCopy} variant="primary" fullWidth>
        {copied ? "Copied!" : "Copy Link"}
      </Button>
      <Button onClick={onClose} variant="secondary" fullWidth>
        Close
      </Button>
    </div>
  </div>
);
