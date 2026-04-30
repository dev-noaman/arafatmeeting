-- meetings
CREATE TABLE meetings (
  id BIGSERIAL PRIMARY KEY,
  meeting_code VARCHAR(12) NOT NULL UNIQUE,
  creator_id TEXT NOT NULL,
  meeting_link TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- meeting attendees
CREATE TABLE meeting_attendees (
  id BIGSERIAL PRIMARY KEY,
  meeting_id BIGINT NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
  user_id TEXT,
  email TEXT NOT NULL,
  name TEXT,
  invited_at TIMESTAMPTZ DEFAULT NOW(),
  joined_at TIMESTAMPTZ
);

-- saved attendees
CREATE TABLE saved_attendees (
  id BIGSERIAL PRIMARY KEY,
  user_id TEXT NOT NULL,
  email TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, email)
);

-- meeting cache (for Go backend ownership checks)
CREATE TABLE meeting_cache (
  meeting_id BIGSERIAL PRIMARY KEY,
  meeting_code VARCHAR(12) NOT NULL UNIQUE,
  creator_id TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- meeting code generation function
CREATE OR REPLACE FUNCTION generate_meeting_code()
RETURNS TEXT AS $$
DECLARE
  chars TEXT := 'abcdefghijklmnopqrstuvwxyz';
  code TEXT;
BEGIN
  code := substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) || '-' ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) || '-' ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1);
  RETURN code;
END;
$$ LANGUAGE plpgsql;

-- RLS
ALTER TABLE meetings ENABLE ROW LEVEL SECURITY;
ALTER TABLE meeting_attendees ENABLE ROW LEVEL SECURITY;
ALTER TABLE saved_attendees ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own meetings" ON meetings FOR SELECT USING (creator_id = auth.uid()::text);
CREATE POLICY "Users can create meetings" ON meetings FOR INSERT WITH CHECK (creator_id = auth.uid()::text);
CREATE POLICY "Users can update own meetings" ON meetings FOR UPDATE USING (creator_id = auth.uid()::text);
CREATE POLICY "Users can delete own meetings" ON meetings FOR DELETE USING (creator_id = auth.uid()::text);
CREATE POLICY "Public meeting code lookup" ON meetings FOR SELECT USING (true);

CREATE POLICY "Meeting attendees visible to creator" ON meeting_attendees FOR SELECT USING (
  EXISTS (SELECT 1 FROM meetings WHERE meetings.id = meeting_attendees.meeting_id AND meetings.creator_id = auth.uid()::text)
  OR user_id = auth.uid()::text
);
CREATE POLICY "Creator can manage attendees" ON meeting_attendees FOR ALL USING (
  EXISTS (SELECT 1 FROM meetings WHERE meetings.id = meeting_attendees.meeting_id AND meetings.creator_id = auth.uid()::text)
);

CREATE POLICY "Users can manage own saved attendees" ON saved_attendees FOR ALL USING (user_id = auth.uid()::text);

-- Indexes
CREATE INDEX idx_meetings_creator ON meetings(creator_id);
CREATE INDEX idx_meetings_code ON meetings(meeting_code);
CREATE INDEX idx_meeting_attendees_meeting ON meeting_attendees(meeting_id);
CREATE INDEX idx_saved_attendees_user ON saved_attendees(user_id);
