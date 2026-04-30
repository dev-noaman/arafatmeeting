import React, { useState } from "react";
import {
  useParticipants,
  useLocalParticipant,
} from "@livekit/components-react";
import {
  removeParticipant,
  muteParticipant,
  endMeeting,
} from "../../../services/api/livekit";
import type { AdminControlsProps } from "./types";
import { EndMeetingButton } from "./EndMeetingButton";
import { ParticipantsList } from "./ParticipantsList";
import { InviteAttendeesModal } from "../InviteAttendeesModal";
import { Button } from "../../common/Button";

export const AdminControls: React.FC<AdminControlsProps> = ({
  meetingCode,
  isAdmin,
  onEndMeeting,
  meetingId,
}) => {
  const [isEndingMeeting, setIsEndingMeeting] = useState(false);
  const [showEndConfirm, setShowEndConfirm] = useState(false);
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const participants = useParticipants();
  const { localParticipant } = useLocalParticipant();

  if (!isAdmin) return null;

  const handleKickParticipant = async (identity: string) => {
    try {
      await removeParticipant(meetingCode, identity);
    } catch (error) {
      console.error("Failed to kick participant:", error);
      alert("Failed to kick participant");
    }
  };

  const handleMuteTrack = async (
    identity: string,
    trackSid: string,
    muted: boolean,
  ) => {
    try {
      await muteParticipant(meetingCode, identity, trackSid, muted);
    } catch (error) {
      console.error("Failed to mute participant:", error);
      alert("Failed to mute participant");
    }
  };

  const handleEndMeeting = async () => {
    setIsEndingMeeting(true);
    try {
      await endMeeting(meetingCode);
      onEndMeeting();
    } catch (error) {
      console.error("Failed to end meeting:", error);
      alert("Failed to end meeting");
      setIsEndingMeeting(false);
    }
  };

  return (
    <>
      <div
        className="flex flex-col h-full overflow-hidden"
        style={{ background: "var(--lk-bg2)" }}
      >
        <div
          className="p-4 shrink-0"
          style={{ borderBottom: "1px solid var(--lk-border-color)" }}
        >
          <Button
            onClick={() => setIsInviteModalOpen(true)}
            variant="primary"
            fullWidth
            className="mb-3"
          >
            <svg
              className="h-5 w-5 inline-block mr-2"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"
              />
            </svg>
            Invite Attendees
          </Button>
          <EndMeetingButton
            onEndMeeting={handleEndMeeting}
            isEndingMeeting={isEndingMeeting}
            showConfirm={showEndConfirm}
            onShowConfirm={setShowEndConfirm}
          />
        </div>

        <ParticipantsList
          participants={
            participants as Array<{
              identity: string;
              name?: string;
              metadata?: string;
            }>
          }
          localParticipantIdentity={localParticipant.identity}
          onKick={handleKickParticipant}
          onMuteTrack={handleMuteTrack}
        />
      </div>

      <InviteAttendeesModal
        isOpen={isInviteModalOpen}
        onClose={() => setIsInviteModalOpen(false)}
        meetingId={meetingId || 0}
      />
    </>
  );
};
