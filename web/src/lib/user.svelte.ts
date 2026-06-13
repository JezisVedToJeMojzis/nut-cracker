// Current-user state, backed by the session cookie. On load we ask the backend
// who we are (/auth/me); pages read `user.id` / `user.username` reactively.
import { getMe, type Profile } from './api';

let me = $state<Profile | null>(null);
let ready = $state(false);

export const user = {
	/** Numeric id as a string (empty when logged out). */
	get id(): string {
		return me ? String(me.id) : '';
	},
	get username(): string {
		return me?.username ?? '';
	},
	get email(): string {
		return me?.email ?? '';
	},
	get profile(): Profile | null {
		return me;
	},
	get isAuthed(): boolean {
		return me !== null;
	},
	/** True once the initial /auth/me check has completed. */
	get ready(): boolean {
		return ready;
	},
	/** Set the current user (after login/register). */
	set(profile: Profile | null) {
		me = profile;
	},
	/** Re-check the session with the backend. */
	async refresh() {
		try {
			me = await getMe();
		} catch {
			me = null;
		}
		ready = true;
	}
};
