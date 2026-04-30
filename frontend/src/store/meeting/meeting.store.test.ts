import { describe, it, expect, beforeEach } from "vitest";
import { useMeetingStore } from "./meeting.store";

describe("useMeetingStore", () => {
  beforeEach(() => {
    useMeetingStore.getState().resetMeeting();
  });

  it("starts with default state", () => {
    const state = useMeetingStore.getState();
    expect(state.meetingCode).toBeNull();
    expect(state.isInMeeting).toBe(false);
    expect(state.isAdmin).toBe(false);
    expect(state.participants).toEqual([]);
    expect(state.isSummarizerActive).toBe(false);
  });

  it("sets meeting code", () => {
    useMeetingStore.getState().setMeetingCode("abc-defg-hijk");
    expect(useMeetingStore.getState().meetingCode).toBe("abc-defg-hijk");
  });

  it("sets isInMeeting flag", () => {
    useMeetingStore.getState().setIsInMeeting(true);
    expect(useMeetingStore.getState().isInMeeting).toBe(true);
  });

  it("sets isAdmin flag", () => {
    useMeetingStore.getState().setIsAdmin(true);
    expect(useMeetingStore.getState().isAdmin).toBe(true);
  });

  it("adds a participant", () => {
    useMeetingStore.getState().addParticipant({
      identity: "user1",
      name: "User 1",
      isCameraOn: true,
      isMicrophoneOn: true,
    });
    expect(useMeetingStore.getState().participants).toHaveLength(1);
    expect(useMeetingStore.getState().participants[0].identity).toBe("user1");
  });

  it("removes a participant by identity", () => {
    useMeetingStore.getState().addParticipant({
      identity: "user1",
      name: "User 1",
      isCameraOn: true,
      isMicrophoneOn: true,
    });
    useMeetingStore.getState().addParticipant({
      identity: "user2",
      name: "User 2",
      isCameraOn: false,
      isMicrophoneOn: true,
    });
    useMeetingStore.getState().removeParticipant("user1");
    expect(useMeetingStore.getState().participants).toHaveLength(1);
    expect(useMeetingStore.getState().participants[0].identity).toBe("user2");
  });

  it("updates a participant", () => {
    useMeetingStore.getState().addParticipant({
      identity: "user1",
      name: "User 1",
      isCameraOn: true,
      isMicrophoneOn: true,
    });
    useMeetingStore.getState().updateParticipant("user1", {
      isMicrophoneOn: false,
    });
    const p = useMeetingStore.getState().participants[0];
    expect(p.isMicrophoneOn).toBe(false);
    expect(p.isCameraOn).toBe(true);
    expect(p.name).toBe("User 1");
  });

  it("sets summarizer active", () => {
    useMeetingStore.getState().setSummarizerActive(true);
    expect(useMeetingStore.getState().isSummarizerActive).toBe(true);
  });

  it("resets meeting state", () => {
    useMeetingStore.getState().setMeetingCode("abc-defg-hijk");
    useMeetingStore.getState().setIsInMeeting(true);
    useMeetingStore.getState().setIsAdmin(true);
    useMeetingStore.getState().addParticipant({
      identity: "user1",
      name: "User 1",
      isCameraOn: true,
      isMicrophoneOn: true,
    });
    useMeetingStore.getState().resetMeeting();
    const state = useMeetingStore.getState();
    expect(state.meetingCode).toBeNull();
    expect(state.isInMeeting).toBe(false);
    expect(state.isAdmin).toBe(false);
    expect(state.participants).toEqual([]);
    expect(state.isSummarizerActive).toBe(false);
  });
});
