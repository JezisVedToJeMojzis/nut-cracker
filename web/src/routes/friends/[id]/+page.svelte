<script lang="ts">
	import { page } from '$app/state';
	import WorldMap from '$lib/WorldMap.svelte';
	import { getMap, listFriends } from '$lib/api';
	import { user } from '$lib/user.svelte';

	let cracks = $state<Record<string, number>>({});
	let username = $state('');
	let error = $state('');
	let loading = $state(false);

	const ownerId = $derived(page.params.id ?? '');

	$effect(() => {
		if (user.id && ownerId) load(user.id, ownerId);
	});

	async function load(viewer: string, owner: string) {
		loading = true;
		error = '';
		try {
			const [list, friends] = await Promise.all([getMap(viewer, owner), listFriends(viewer)]);
			const next: Record<string, number> = {};
			for (const c of list) next[c.country_code] = c.cracks;
			cracks = next;
			username = friends.find((f) => f.id === owner)?.username ?? owner;
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	const total = $derived(Object.keys(cracks).length);
	const noop = () => {};
</script>

<a href="/friends" class="back">← Back to friends</a>
<h1>{username ? `${username}'s map` : 'Map'}</h1>

{#if error}
	<p class="error">⚠️ {error}</p>
	<p class="hint">You can only view the maps of users who are your accepted friends.</p>
{/if}
{#if loading}<p>Loading…</p>{/if}
{#if !error}<p class="hint"><strong>{total}</strong> countries cracked (view only).</p>{/if}

<WorldMap {cracks} countMode={true} oncrack={noop} onuncrack={noop} />

<style>
	h1 {
		margin: 0.25rem 0 0.5rem;
	}
	.back {
		color: #6b7280;
		text-decoration: none;
		font-size: 0.9rem;
	}
	.hint {
		color: #6b7280;
		font-size: 0.9rem;
	}
	.error {
		color: #b91c1c;
	}
</style>
