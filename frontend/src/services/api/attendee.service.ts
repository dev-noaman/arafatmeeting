import { insforge } from "../insforge/client";
import type { AttendeeResponse, SavedAttendee } from "../../types/meeting.types";

export const attendeeService = {
  addAttendee: async (meetingId: number, email: string, name?: string): Promise<AttendeeResponse> => {
    const { data, error } = await insforge.database
      .from("meeting_attendees")
      .insert({ meeting_id: meetingId, email, name })
      .select()
      .single();
    if (error) throw new Error(error.message);
    return data as AttendeeResponse;
  },

  removeAttendee: async (meetingId: number, email: string): Promise<void> => {
    const { error } = await insforge.database
      .from("meeting_attendees")
      .delete()
      .eq("meeting_id", meetingId)
      .eq("email", email);
    if (error) throw new Error(error.message);
  },

  getSavedAttendees: async (): Promise<SavedAttendee[]> => {
    const { data, error } = await insforge.database
      .from("saved_attendees")
      .select("*")
      .order("created_at", { ascending: false });
    if (error) throw new Error(error.message);
    return (data as SavedAttendee[]) || [];
  },

  saveAttendee: async (email: string, name: string): Promise<SavedAttendee> => {
    const { data: userData } = await insforge.auth.getCurrentUser();
    const userId = userData.user?.id;
    const { data, error } = await insforge.database
      .from("saved_attendees")
      .insert({ user_id: userId, email, name })
      .select()
      .single();
    if (error) throw new Error(error.message);
    return data as SavedAttendee;
  },

  updateAttendee: async (id: number, email: string, name: string): Promise<SavedAttendee> => {
    const { data, error } = await insforge.database
      .from("saved_attendees")
      .update({ email, name })
      .eq("id", id)
      .select()
      .single();
    if (error) throw new Error(error.message);
    return data as SavedAttendee;
  },

  deleteAttendee: async (id: number): Promise<void> => {
    const { error } = await insforge.database.from("saved_attendees").delete().eq("id", id);
    if (error) throw new Error(error.message);
  },
};
