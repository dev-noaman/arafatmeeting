import { DisconnectReason } from "livekit-client";

export function parseTokenMetadata(token: string): { isAdmin: boolean } {
  if (!token) return { isAdmin: false };

  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    const metadata = payload.metadata ? JSON.parse(payload.metadata) : {};
    return { isAdmin: metadata.role === "admin" };
  } catch (e) {
    console.error("Failed to parse token metadata:", e);
    return { isAdmin: false };
  }
}

export function getDisconnectMessage(reason?: DisconnectReason): string {
  if (reason === DisconnectReason.SERVER_SHUTDOWN) {
    return "Meeting ended by host";
  } else if (reason === DisconnectReason.PARTICIPANT_REMOVED) {
    return "You were removed from the meeting";
  } else if (reason === DisconnectReason.ROOM_DELETED) {
    return "Meeting ended by host";
  } else if (reason === DisconnectReason.CLIENT_INITIATED) {
    return "You left the meeting";
  } else if (reason === DisconnectReason.DUPLICATE_IDENTITY) {
    return "You joined from another device";
  }
  // Default: connection failed or unknown reason
  return "Connection to meeting ended";
}
