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

<div class="head">
	<div>
		<h1>{username ? `${username}'s map` : 'Map'}</h1>
		<p class="muted sub">View only · you can see this because you're friends.</p>
	</div>
	{#if !error}
		<div class="stat">
			<span class="num">{total}</span>
			<span class="muted">cracked</span>
		</div>
	{/if}
</div>

{#if error}
	<div class="card empty">
		<span class="emoji">🔒</span>
		<p>{error}</p>
		<p class="muted small">You can only view the maps of your accepted friends.</p>
	</div>
{:else}
	{#if loading}<p class="muted"><span class="spinner"></span> Loading…</p>{/if}
	<WorldMap {cracks} countMode={true} oncrack={noop} onuncrack={noop} />
{/if}

<style>
	.back {
		color: var(--muted);
		text-decoration: none;
		font-size: 0.88rem;
	}
	.back:hover {
		color: var(--text);
	}
	.head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin: 0.4rem 0 1rem;
	}
	h1 {
		margin: 0;
	}
	.sub {
		margin: 0.25rem 0 0;
		font-size: 0.9rem;
	}
	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 0.5rem 1rem;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		box-shadow: var(--shadow-sm);
		line-height: 1.1;
	}
	.stat .num {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--primary-700);
	}
	.stat .muted {
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.4rem;
		padding: 2.5rem;
		text-align: center;
	}
	.empty .emoji {
		font-size: 2.2rem;
	}
	.small {
		font-size: 0.82rem;
	}
</style>
