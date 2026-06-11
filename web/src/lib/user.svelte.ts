// Shared current-user state (temporary auth stand-in).
// Holds the user UUID sent as X-User-ID and persists it in localStorage.
import { browser } from '$app/environment';

const KEY = 'nutcracker_user_id';

let _id = $state(browser ? (localStorage.getItem(KEY) ?? '') : '');

export const user = {
	get id(): string {
		return _id;
	},
	set id(value: string) {
		_id = value;
		if (browser) localStorage.setItem(KEY, value);
	}
};
