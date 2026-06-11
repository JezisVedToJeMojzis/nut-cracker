// Minimal client for the Nut Cracker backend API.
// Requests go through the Vite proxy (/api -> http://localhost:8080).
//
// Auth is not built yet: the acting user is identified by the X-User-ID header,
// a temporary stand-in to be replaced by real sessions/OAuth later.

export type CountryCount = {
	country_code: string;
	cracks: number;
};

export type IncrementResult = {
	country_code: string;
	cracks: number;
};

export type DecrementResult = {
	country_code: string;
	cracks: number;
	removed: boolean;
};

export type Settings = {
	count_mode: boolean;
};

export type Friend = {
	id: number;
	username: string;
};

export type Profile = {
	id: number;
	username: string;
	email: string;
	created_at: string;
};

export type UserCard = {
	id: number;
	username: string;
};

const BASE = '/api';

function headers(userId: string): HeadersInit {
	return {
		'Content-Type': 'application/json',
		'X-User-ID': userId
	};
}

async function handle<T>(res: Response): Promise<T> {
	if (!res.ok) {
		let msg = `request failed (${res.status})`;
		try {
			const body = await res.json();
			if (body?.error) msg = body.error;
		} catch {
			/* ignore non-JSON bodies */
		}
		throw new Error(msg);
	}
	return res.json() as Promise<T>;
}

/** Fetch a user's map of cracked countries. */
export async function getMap(userId: string, ownerId: string): Promise<CountryCount[]> {
	const res = await fetch(`${BASE}/users/${ownerId}/map`, { headers: headers(userId) });
	const data = await handle<{ countries: CountryCount[] }>(res);
	return data.countries;
}

/** Add one crack to a country on the caller's own map. */
export async function increment(userId: string, code: string): Promise<IncrementResult> {
	const res = await fetch(`${BASE}/users/${userId}/countries/${code}/increment`, {
		method: 'POST',
		headers: headers(userId)
	});
	return handle<IncrementResult>(res);
}

/** Remove one crack; the country is removed from the map when it reaches zero. */
export async function decrement(userId: string, code: string): Promise<DecrementResult> {
	const res = await fetch(`${BASE}/users/${userId}/countries/${code}/decrement`, {
		method: 'POST',
		headers: headers(userId)
	});
	return handle<DecrementResult>(res);
}

/** Fetch the caller's own full profile. */
export async function getProfile(userId: string): Promise<Profile> {
	const res = await fetch(`${BASE}/users/${userId}`, { headers: headers(userId) });
	return handle<Profile>(res);
}

/** Update the caller's username. */
export async function updateUsername(userId: string, username: string): Promise<Profile> {
	const res = await fetch(`${BASE}/users/${userId}`, {
		method: 'PATCH',
		headers: headers(userId),
		body: JSON.stringify({ username })
	});
	return handle<Profile>(res);
}

/** Look up a user's public card (id + username) by their numeric id. */
export async function getCard(userId: string, targetId: number): Promise<UserCard> {
	const res = await fetch(`${BASE}/users/${targetId}/card`, { headers: headers(userId) });
	return handle<UserCard>(res);
}

/** List the caller's accepted friends. */
export async function listFriends(userId: string): Promise<Friend[]> {
	const res = await fetch(`${BASE}/friends`, { headers: headers(userId) });
	const data = await handle<{ friends: Friend[] }>(res);
	return data.friends;
}

/** List pending requests received by the caller. */
export async function listIncoming(userId: string): Promise<Friend[]> {
	const res = await fetch(`${BASE}/friends/requests/incoming`, { headers: headers(userId) });
	const data = await handle<{ requests: Friend[] }>(res);
	return data.requests;
}

/** List pending requests sent by the caller. */
export async function listOutgoing(userId: string): Promise<Friend[]> {
	const res = await fetch(`${BASE}/friends/requests/outgoing`, { headers: headers(userId) });
	const data = await handle<{ requests: Friend[] }>(res);
	return data.requests;
}

/** Send a friend request to another user by numeric id. */
export async function sendRequest(userId: string, to: number): Promise<{ status: string }> {
	const res = await fetch(`${BASE}/friends/requests`, {
		method: 'POST',
		headers: headers(userId),
		body: JSON.stringify({ to })
	});
	return handle<{ status: string }>(res);
}

/** Accept a pending request from requesterId. */
export async function acceptRequest(userId: string, requesterId: number): Promise<void> {
	const res = await fetch(`${BASE}/friends/requests/${requesterId}/accept`, {
		method: 'POST',
		headers: headers(userId)
	});
	await handle<unknown>(res);
}

/** Decline a pending request from requesterId. */
export async function declineRequest(userId: string, requesterId: number): Promise<void> {
	const res = await fetch(`${BASE}/friends/requests/${requesterId}/decline`, {
		method: 'POST',
		headers: headers(userId)
	});
	await handle<unknown>(res);
}

/** Remove a relationship (unfriend / cancel / decline) with otherId. */
export async function removeFriend(userId: string, otherId: number): Promise<void> {
	const res = await fetch(`${BASE}/friends/${otherId}`, {
		method: 'DELETE',
		headers: headers(userId)
	});
	if (!res.ok && res.status !== 204) throw new Error(`request failed (${res.status})`);
}

/** Fetch the caller's settings (feature flags). */
export async function getSettings(userId: string): Promise<Settings> {
	const res = await fetch(`${BASE}/users/${userId}/settings`, { headers: headers(userId) });
	return handle<Settings>(res);
}

/** Update the caller's settings. */
export async function updateSettings(userId: string, settings: Settings): Promise<Settings> {
	const res = await fetch(`${BASE}/users/${userId}/settings`, {
		method: 'PUT',
		headers: headers(userId),
		body: JSON.stringify(settings)
	});
	return handle<Settings>(res);
}

/** Remove a country from the caller's own map entirely. */
export async function remove(userId: string, code: string): Promise<void> {
	const res = await fetch(`${BASE}/users/${userId}/countries/${code}`, {
		method: 'DELETE',
		headers: headers(userId)
	});
	if (!res.ok && res.status !== 204) {
		throw new Error(`request failed (${res.status})`);
	}
}
