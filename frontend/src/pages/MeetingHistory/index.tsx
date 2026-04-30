import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Layout } from "../../components/layout/Layout";
import { meetingService } from "../../services/api/meeting.service";
import type { Meeting, SummarizerSession } from "../../types/meeting.types";

interface MeetingWithSession extends Meeting {
  session?: SummarizerSession;
}

export default function MeetingHistory() {
  const navigate = useNavigate();
  const [meetings, setMeetings] = useState<MeetingWithSession[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [copiedLinkId, setCopiedLinkId] = useState<number | null>(null);

  useEffect(() => {
    loadMeetings();
  }, []);

  const loadMeetings = async () => {
    try {
      setIsLoading(true);
      const userMeetings = await meetingService.getMyMeetings();

      // Fetch sessions for each meeting
      const meetingsWithSessions = await Promise.all(
        userMeetings.map(async (meeting) => {
          try {
            const sessions = await meetingService.getSummarizerSessions(meeting.id);
            const latestSession = sessions.length > 0 ? sessions[sessions.length - 1] : undefined;
            return { ...meeting, session: latestSession };
          } catch {
            return meeting;
          }
        })
      );

      setMeetings(meetingsWithSessions);
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to load meeting history");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteMeeting = async (meetingId: number) => {
    if (!confirm("Are you sure you want to delete this meeting? This action cannot be undone.")) {
      return;
    }

    try {
      setDeletingId(meetingId);
      await meetingService.deleteMeeting(meetingId);
      setMeetings(meetings.filter((m) => m.id !== meetingId));
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to delete meeting");
    } finally {
      setDeletingId(null);
    }
  };

  const handleJoinMeeting = (meetingCode: string) => {
    navigate(`/${meetingCode}`);
  };

  const handleCopyMeetingLink = async (meeting: MeetingWithSession) => {
    try {
      const fullLink = `${window.location.origin}/${meeting.meeting_code}`;
      await navigator.clipboard.writeText(fullLink);
      setCopiedLinkId(meeting.id);
      setTimeout(() => setCopiedLinkId(null), 2000);
    } catch {
      const textArea = document.createElement("textarea");
      textArea.value = `${window.location.origin}/${meeting.meeting_code}`;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand("copy");
      document.body.removeChild(textArea);
      setCopiedLinkId(meeting.id);
      setTimeout(() => setCopiedLinkId(null), 2000);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  };

  if (isLoading) {
    return (
      <Layout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex justify-center items-center h-64">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-brand-600"></div>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Meeting History</h1>
          <p className="mt-2 text-gray-600">
            View and manage your past meetings, summaries, and transcripts
          </p>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            {error}
          </div>
        )}

        {meetings.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg shadow">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No meetings yet</h3>
            <p className="mt-1 text-sm text-gray-500">
              Create your first meeting to see it here
            </p>
            <div className="mt-6">
              <button
                onClick={() => navigate("/")}
                className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-brand-600 hover:bg-brand-700"
              >
                Create Meeting
              </button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            {meetings.map((meeting) => (
              <div
                key={meeting.id}
                className="bg-white rounded-lg shadow p-6 hover:shadow-md transition-shadow"
              >
                {/* Meeting Header */}
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <h3 className="text-lg font-semibold text-gray-900">
                        Meeting #{meeting.meeting_code}
                      </h3>
                      {meeting.session && (
                        <span
                          className={`px-2 py-1 text-xs font-medium rounded-full ${
                            meeting.session.status === "SUMMARIZED"
                              ? "bg-green-100 text-green-800"
                              : meeting.session.status === "FAILED"
                              ? "bg-red-100 text-red-800"
                              : "bg-yellow-100 text-yellow-800"
                          }`}
                        >
                          {meeting.session.status}
                        </span>
                      )}
                    </div>
                    <p className="mt-1 text-sm text-gray-500">
                      Created {formatDate(meeting.created_at)}
                    </p>
                  </div>

                  {/* Action Buttons */}
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleCopyMeetingLink(meeting)}
                      className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded transition-colors"
                      title="Copy meeting link"
                    >
                      {copiedLinkId === meeting.id ? "Copied!" : "Copy Link"}
                    </button>
                    <button
                      onClick={() => handleJoinMeeting(meeting.meeting_code)}
                      className="px-3 py-1.5 text-sm bg-brand-600 hover:bg-brand-700 text-white rounded transition-colors"
                    >
                      Join
                    </button>
                    <button
                      onClick={() => handleDeleteMeeting(meeting.id)}
                      disabled={deletingId === meeting.id}
                      className="px-3 py-1.5 text-sm bg-red-600 hover:bg-red-700 disabled:bg-gray-400 text-white rounded transition-colors"
                    >
                      {deletingId === meeting.id ? "Deleting..." : "Delete"}
                    </button>
                  </div>
                </div>

                {/* Session Info */}
                {meeting.session && (
                  <div className="mt-4 border-t pt-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="text-sm font-medium text-gray-900">
                        {meeting.session.status === "SUMMARIZED"
                          ? "Summary Available"
                          : "Processing..."}
                      </h4>
                    </div>

                    {meeting.session.transcript && (
                      <div className="mt-2">
                        <p className="text-xs text-gray-500 mb-1">Transcript:</p>
                        <div className="p-3 bg-gray-50 rounded text-sm text-gray-700 max-h-32 overflow-y-auto">
                          {meeting.session.transcript.length > 200
                            ? `${meeting.session.transcript.substring(0, 200)}...`
                            : meeting.session.transcript}
                        </div>
                      </div>
                    )}

                    {meeting.session.summary && (
                      <div className="mt-2">
                        <p className="text-xs text-gray-500 mb-1">Summary:</p>
                        <div className="p-3 bg-green-50 rounded text-sm text-gray-700 max-h-32 overflow-y-auto">
                          {meeting.session.summary.length > 300
                            ? `${meeting.session.summary.substring(0, 300)}...`
                            : meeting.session.summary}
                        </div>
                      </div>
                    )}

                    {meeting.session.error && (
                      <div className="mt-2 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                        Error: {meeting.session.error}
                      </div>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </Layout>
  );
}
