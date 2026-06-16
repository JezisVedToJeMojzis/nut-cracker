// Client for the Nut Cracker backend API.
// Requests go through the Vite proxy (/api -> backend) and rely on the
// HttpOnly session cookie for authentication (credentials: 'include').

export type CountryCount = { country_code: string; cracks: number };
export type IncrementResult = { country_code: string; cracks: number };
export type DecrementResult = { country_code: string; cracks: number; removed: boolean };
export type Settings = { count_mode: boolean };
export type Friend = { id: number; username: string };
export type UserCard = { id: number; username: string };
export type Profile = {
	id: number;
	username: string;
	email: string;
	created_at: string;
};

const BASE = '/api';

type Options = { method?: string; body?: unknown };

async function req<T>(path: string, opts: Options = {}): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method: opts.method ?? 'GET',
		credentials: 'include',
		headers: opts.body ? { 'Content-Type': 'application/json' } : undefined,
		body: opts.body ? JSON.stringify(opts.body) : undefined
	});
	if (res.status === 204) return undefined as T;
	if (!res.ok) {
		let msg = `request failed (${res.status})`;
		try {
			const b = await res.json();
			if (b?.error) msg = b.error;
		} catch {
			/* ignore */
		}
		throw new Error(msg);
	}
	return res.json() as Promise<T>;
}

// --- Auth ----------------------------------------------------------------

export const register = (email: string, username: string, password: string) =>
	req<Profile>('/auth/register', { method: 'POST', body: { email, username, password } });

export const login = (email: string, password: string) =>
	req<Profile>('/auth/login', { method: 'POST', body: { email, password } });

export const logout = () => req<{ ok: boolean }>('/auth/logout', { method: 'POST' });

export const getMe = () => req<Profile>('/auth/me');

// --- Profile -------------------------------------------------------------

export const updateUsername = (userId: string, username: string) =>
	req<Profile>(`/users/${userId}`, { method: 'PATCH', body: { username } });

export const getCard = (targetId: number) => req<UserCard>(`/users/${targetId}/card`);

// --- Map -----------------------------------------------------------------

export async function getMap(ownerId: string): Promise<CountryCount[]> {
	const data = await req<{ countries: CountryCount[] }>(`/users/${ownerId}/map`);
	return data.countries;
}

export const increment = (userId: string, code: string) =>
	req<IncrementResult>(`/users/${userId}/countries/${code}/increment`, { method: 'POST' });

export const decrement = (userId: string, code: string) =>
	req<DecrementResult>(`/users/${userId}/countries/${code}/decrement`, { method: 'POST' });

export const remove = (userId: string, code: string) =>
	req<void>(`/users/${userId}/countries/${code}`, { method: 'DELETE' });

// --- Settings ------------------------------------------------------------

export const getSettings = (userId: string) => req<Settings>(`/users/${userId}/settings`);

export const updateSettings = (userId: string, settings: Settings) =>
	req<Settings>(`/users/${userId}/settings`, { method: 'PUT', body: settings });

// --- Friends -------------------------------------------------------------

export async function listFriends(): Promise<Friend[]> {
	return (await req<{ friends: Friend[] }>('/friends')).friends;
}
export async function listIncoming(): Promise<Friend[]> {
	return (await req<{ requests: Friend[] }>('/friends/requests/incoming')).requests;
}
export async function listOutgoing(): Promise<Friend[]> {
	return (await req<{ requests: Friend[] }>('/friends/requests/outgoing')).requests;
}
export const sendRequest = (to: number) =>
	req<{ status: string }>('/friends/requests', { method: 'POST', body: { to } });

export const acceptRequest = (requesterId: number) =>
	req<unknown>(`/friends/requests/${requesterId}/accept`, { method: 'POST' });

export const declineRequest = (requesterId: number) =>
	req<unknown>(`/friends/requests/${requesterId}/decline`, { method: 'POST' });

export const removeFriend = (otherId: number) =>
	req<void>(`/friends/${otherId}`, { method: 'DELETE' });
