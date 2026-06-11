// Shared current-user state (temporary auth stand-in).
// Holds the user's numeric id (sent as X-User-ID) and caches the username for
// display. Persisted in localStorage.
import { browser } from '$app/environment';

const ID_KEY = 'nutcracker_user_id';
const NAME_KEY = 'nutcracker_username';

let _id = $state(browser ? (localStorage.getItem(ID_KEY) ?? '') : '');
let _username = $state(browser ? (localStorage.getItem(NAME_KEY) ?? '') : '');

export const user = {
	get id(): string {
		return _id;
	},
	set id(value: string) {
		_id = value;
		if (browser) localStorage.setItem(ID_KEY, value);
	},
	get username(): string {
		return _username;
	},
	set username(value: string) {
		_username = value;
		if (browser) localStorage.setItem(NAME_KEY, value);
	},
	signOut() {
		_id = '';
		_username = '';
		if (browser) {
			localStorage.removeItem(ID_KEY);
			localStorage.removeItem(NAME_KEY);
		}
	}
};
