import { insforge } from "../insforge/client";
import type {
  Meeting,
  SummarizerSession,
  SummarizerStartResponse,
  SummarizerStopResponse,
} from "../../types/meeting.types";
import apiClient from "./client";

export const meetingService = {
  createMeeting: async (attendees?: string[]): Promise<Meeting> => {
    const { data: userData, error: userError } = await insforge.auth.getCurrentUser();
    if (userError || !userData.user) throw new Error("Not authenticated");

    const { data: codeResult, error: codeError } = await insforge.database.rpc("generate_meeting_code");
    if (codeError) throw new Error(codeError.message);

    const { data, error } = await insforge.database
      .from("meetings")
      .insert({ meeting_code: codeResult, creator_id: userData.user.id })
      .select()
      .single();
    if (error) throw new Error(error.message);

    return data as Meeting;
  },

  getMyMeetings: async (): Promise<Meeting[]> => {
    const { data: userData } = await insforge.auth.getCurrentUser();
    const userId = userData.user?.id;
    if (!userId) return [];

    const { data, error } = await insforge.database
      .from("meetings")
      .select("*")
      .eq("creator_id", userId)
      .order("created_at", { ascending: false });
    if (error) throw new Error(error.message);
    return (data as Meeting[]) || [];
  },

  getMeetingById: async (id: number): Promise<Meeting> => {
    const { data, error } = await insforge.database
      .from("meetings")
      .select("*")
      .eq("id", id)
      .single();
    if (error) throw new Error(error.message);
    return data as Meeting;
  },

  getMeetingByCode: async (code: string): Promise<Meeting> => {
    const { data, error } = await insforge.database
      .from("meetings")
      .select("*")
      .eq("meeting_code", code)
      .single();
    if (error) throw new Error(error.message);
    return data as Meeting;
  },

  deleteMeeting: async (id: number): Promise<void> => {
    const { error } = await insforge.database.from("meetings").delete().eq("id", id);
    if (error) throw new Error(error.message);
  },

  getAllMeetings: async (): Promise<Meeting[]> => {
    const { data, error } = await insforge.database
      .from("meetings")
      .select("*")
      .order("created_at", { ascending: false });
    if (error) throw new Error(error.message);
    return (data as Meeting[]) || [];
  },

  // Summarizer operations stay on Go backend
  startSummarizer: async (meetingId: number): Promise<SummarizerSession> => {
    const response = await apiClient.post<SummarizerStartResponse>(`/meetings/${meetingId}/summarizer/start`);
    return response.data.data;
  },

  stopSummarizer: async (meetingId: number): Promise<SummarizerSession> => {
    const response = await apiClient.post<SummarizerStopResponse>(`/meetings/${meetingId}/summarizer/stop`);
    return response.data.data;
  },

  getSummarizerSessions: async (meetingId: number): Promise<SummarizerSession[]> => {
    const response = await apiClient.get(`/meetings/${meetingId}/summarizer/sessions`);
    return response.data.data;
  },
};
