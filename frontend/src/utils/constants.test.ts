import { describe, it, expect } from "vitest";
import { COLORS, BREAKPOINTS, VIDEO_CONSTRAINTS, ERROR_MESSAGES, A11Y } from "./constants";

describe("constants", () => {
  it("COLORS has expected shape", () => {
    expect(COLORS.primary).toBe("#1976d2");
    expect(COLORS.error).toBe("#f44336");
    expect(COLORS.success).toBe("#4caf50");
    expect(COLORS.text.primary).toBe("#212121");
    expect(COLORS.text.secondary).toBe("#757575");
  });

  it("BREAKPOINTS has mobile, tablet, desktop", () => {
    expect(BREAKPOINTS.mobile).toBe(640);
    expect(BREAKPOINTS.tablet).toBe(1024);
    expect(BREAKPOINTS.desktop).toBe(1024);
  });

  it("VIDEO_CONSTRAINTS has lobby and meeting presets", () => {
    expect(VIDEO_CONSTRAINTS.lobby.width.ideal).toBe(640);
    expect(VIDEO_CONSTRAINTS.meeting.width.ideal).toBe(1280);
    expect(VIDEO_CONSTRAINTS.meeting.height.ideal).toBe(720);
  });

  it("ERROR_MESSAGES has all expected keys", () => {
    expect(ERROR_MESSAGES.INVALID_MEETING_CODE).toContain("Meeting not found");
    expect(ERROR_MESSAGES.DEVICE_ACCESS_DENIED).toContain("Camera and microphone");
    expect(ERROR_MESSAGES.NETWORK_ERROR).toContain("Connection failed");
    expect(ERROR_MESSAGES.TOKEN_GENERATION_FAILED).toContain("Failed to join");
  });

  it("A11Y has keyboard shortcuts", () => {
    expect(A11Y.keyboardShortcuts.toggleCamera).toBe("v");
    expect(A11Y.keyboardShortcuts.toggleMic).toBe("m");
    expect(A11Y.keyboardShortcuts.leaveMeeting).toBe("Escape");
  });

  it("A11Y has accessibility flags", () => {
    expect(A11Y.focusTrapEnabled).toBe(true);
    expect(A11Y.respectReducedMotion).toBe(true);
  });
});
