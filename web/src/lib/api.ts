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
