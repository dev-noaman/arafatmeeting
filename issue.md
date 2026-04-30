12:32:36.654 Running build in Washington, D.C., USA (East) – iad1
12:32:36.655 Build machine configuration: 2 cores, 8 GB
12:32:36.765 Cloning github.com/dev-noaman/arafatmeeting (Branch: main, Commit: 0ef1a84)
12:32:36.766 Previous build caches not available.
12:32:37.193 Cloning completed: 428.000ms
12:32:37.263 Found .vercelignore
12:32:37.269 Removed 25 ignored files defined in .vercelignore
12:32:37.269   /.env.local
12:32:37.269   /.git/config
12:32:37.269   /.git/description
12:32:37.270   /.git/FETCH_HEAD
12:32:37.270   /.git/HEAD
12:32:37.270   /.git/hooks/applypatch-msg.sample
12:32:37.270   /.git/hooks/commit-msg.sample
12:32:37.270   /.git/hooks/fsmonitor-watchman.sample
12:32:37.270   /.git/hooks/post-update.sample
12:32:37.270   /.git/hooks/pre-applypatch.sample
12:32:37.540 Running "vercel build"
12:32:38.284 Vercel CLI 51.6.1
12:32:38.962 Installing dependencies...
12:32:46.726 
12:32:46.726 added 357 packages in 8s
12:32:46.727 
12:32:46.728 82 packages are looking for funding
12:32:46.728   run `npm fund` for details
12:32:46.779 Running "npm run build"
12:32:46.881 
12:32:46.882 > frontend@0.0.0 build
12:32:46.882 > tsc -b && vite build
12:32:46.882 
12:32:53.347 src/hooks/useApi.test.ts(6,33): error TS2304: Cannot find name 'vi'.
12:32:53.348 src/hooks/useApi.test.ts(9,14): error TS2304: Cannot find name 'vi'.
12:32:53.349 src/pages/AdminUsers/UserRow.tsx(51,21): error TS2367: This comparison appears to be unintentional because the types 'string' and 'number | undefined' have no overlap.
12:32:53.349 src/pages/AdminUsers/index.tsx(35,21): error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'number | undefined'.
12:32:53.349   Type 'string' is not assignable to type 'number'.
12:32:53.350 src/pages/AdminUsers/index.tsx(66,13): error TS2322: Type 'string | undefined' is not assignable to type 'number | undefined'.
12:32:53.350   Type 'string' is not assignable to type 'number'.
12:32:53.350 src/pages/AdminUsers/useDeleteUser.ts(24,9): error TS2367: This comparison appears to be unintentional because the types 'string' and 'number | undefined' have no overlap.
12:32:53.351 src/services/api/lobby.service.test.ts(1,32): error TS6133: 'vi' is declared but its value is never read.
12:32:53.351 src/services/api/lobby.service.test.ts(1,36): error TS6133: 'beforeEach' is declared but its value is never read.
12:32:53.351 src/services/api/meeting.service.ts(11,25): error TS6133: 'attendees' is declared but its value is never read.
12:32:53.352 src/services/api/user/user-admin.service.ts(44,19): error TS2339: Property 'name' does not exist on type '{ id: string; profile: objectOutputType<{ name: ZodOptional<ZodString>; avatar_url: ZodOptional<ZodString>; }, ZodTypeAny, "passthrough"> | null; }'.
12:32:53.353 src/services/api/user/user-admin.service.ts(47,25): error TS2339: Property 'createdAt' does not exist on type '{ id: string; profile: objectOutputType<{ name: ZodOptional<ZodString>; avatar_url: ZodOptional<ZodString>; }, ZodTypeAny, "passthrough"> | null; }'.
12:32:53.353 src/services/api/user/user-profile.service.ts(24,13): error TS6133: 'data' is declared but its value is never read.
12:32:53.353 src/store/auth/authActions.test.ts(31,39): error TS2345: Argument of type 'Mock<Procedure | Constructable>' is not assignable to parameter of type '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>) => void'.
12:32:53.354   Type 'MockInstance<Procedure | Constructable> & (new (...args: any[]) => any) & {}' is not assignable to type '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>) => void'.
12:32:53.354     Type 'MockInstance<Procedure | Constructable> & (new (...args: any[]) => any) & {}' provides no match for the signature '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>): void'.
12:32:53.354 src/store/auth/authActions.test.ts(38,39): error TS2345: Argument of type 'Mock<Procedure | Constructable>' is not assignable to parameter of type '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>) => void'.
12:32:53.354   Type 'MockInstance<Procedure | Constructable> & (new (...args: any[]) => any) & {}' is not assignable to type '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>) => void'.
12:32:53.354     Type 'MockInstance<Procedure | Constructable> & (new (...args: any[]) => any) & {}' provides no match for the signature '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>): void'.
12:32:53.355 src/store/auth/authActions.test.ts(44,39): error TS2345: Argument of type 'Mock<Procedure | Constructable>' is not assignable to parameter of type '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>) => void'.
12:32:53.356   Type 'MockInstance<Procedure | Constructable> & (new (...args: any[]) => any) & {}' is not assignable to type '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>) => void'.
12:32:53.356     Type 'MockInstance<Procedure | Constructable> & (new (...args: any[]) => any) & {}' provides no match for the signature '(partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean; }>): void'.
12:32:53.358 src/store/auth/authActions.ts(19,3): error TS6133: 'get' is declared but its value is never read.
12:32:53.358 src/store/meeting/meeting.store.test.ts(37,7): error TS2353: Object literal may only specify known properties, and 'isCameraOn' does not exist in type 'MeetingParticipant'.
12:32:53.358 src/store/meeting/meeting.store.test.ts(48,7): error TS2353: Object literal may only specify known properties, and 'isCameraOn' does not exist in type 'MeetingParticipant'.
12:32:53.358 src/store/meeting/meeting.store.test.ts(54,7): error TS2353: Object literal may only specify known properties, and 'isCameraOn' does not exist in type 'MeetingParticipant'.
12:32:53.358 src/store/meeting/meeting.store.test.ts(66,7): error TS2353: Object literal may only specify known properties, and 'isCameraOn' does not exist in type 'MeetingParticipant'.
12:32:53.358 src/store/meeting/meeting.store.test.ts(70,7): error TS2353: Object literal may only specify known properties, and 'isMicrophoneOn' does not exist in type 'Partial<MeetingParticipant>'.
12:32:53.358 src/store/meeting/meeting.store.test.ts(73,14): error TS2339: Property 'isMicrophoneOn' does not exist on type 'MeetingParticipant'.
12:32:53.358 src/store/meeting/meeting.store.test.ts(74,14): error TS2339: Property 'isCameraOn' does not exist on type 'MeetingParticipant'.
12:32:53.360 src/store/meeting/meeting.store.test.ts(90,7): error TS2353: Object literal may only specify known properties, and 'isCameraOn' does not exist in type 'MeetingParticipant'.
12:32:53.360 src/store/ui/ui.store.test.ts(10,3): error TS2304: Cannot find name 'afterEach'.
12:32:54.146 vite.config.ts(6,3): error TS2769: No overload matches this call.
12:32:54.146   The last overload gave the following error.
12:32:54.146     Object literal may only specify known properties, and 'test' does not exist in type 'UserConfigExport'.
12:32:54.192 Error: Command "npm run build" exited with 1