<script lang="ts">
	import WorldMap from '$lib/WorldMap.svelte';
	import MapStats from '$lib/MapStats.svelte';
	import { getMap, increment, decrement, getSettings } from '$lib/api';
	import { user } from '$lib/user.svelte';
	import { toaster } from '$lib/toast.svelte';

	let countMode = $state(false);
	let cracks = $state<Record<string, number>>({});
	let loading = $state(false);

	// Reload whenever the current user changes.
	$effect(() => {
		const id = user.id;
		if (id) loadMap(id);
		else {
			cracks = {};
		}
	});

	async function loadMap(id: string) {
		loading = true;
		try {
			const [list, st] = await Promise.all([getMap(id), getSettings(id)]);
			const next: Record<string, number> = {};
			for (const c of list) next[c.country_code] = c.cracks;
			cracks = next;
			countMode = st.count_mode;
		} catch (e) {
			toaster.error(e instanceof Error ? e.message : String(e));
		} finally {
			loading = false;
		}
	}

	async function onCrack(code: string) {
		if (!user.id) {
			toaster.error('Set your user ID first');
			return;
		}
		// With count mode off, an already-coloured country stays as-is.
		if (!countMode && code in cracks) return;
		try {
			const res = await increment(user.id, code);
			cracks = { ...cracks, [code]: res.cracks };
		} catch (e) {
			toaster.error(e instanceof Error ? e.message : String(e));
		}
	}

	async function onUncrack(code: string) {
		if (!user.id || !(code in cracks)) return;
		try {
			const res = await decrement(user.id, code);
			if (res.removed) {
				const next = { ...cracks };
				delete next[code];
				cracks = next;
			} else {
				cracks = { ...cracks, [code]: res.cracks };
			}
		} catch (e) {
			toaster.error(e instanceof Error ? e.message : String(e));
		}
	}
</script>

<div class="head">
	<div>
		<h1>My Map</h1>
		<p class="muted sub">Click a country to mark where you've cracked nuts with someone.</p>
	</div>
</div>

{#if !user.id}
	<div class="card empty">
		<span class="emoji">🌍</span>
		<p>Loading your map…</p>
	</div>
{:else}
	<MapStats {cracks} {countMode} />
	<p class="instructions muted small">
		{#if countMode}
			<strong>Count mode on</strong> · tap a country to add a crack, long-press (or right-click) to
			remove one.
		{:else}
			Tap a country to mark it, long-press (or right-click) to remove it.
		{/if}
		Drag to pan, pinch or scroll to zoom.
		{#if loading}<span class="spinner"></span>{/if}
	</p>
{/if}

<WorldMap {cracks} {countMode} oncrack={onCrack} onuncrack={onUncrack} />

<style>
	.head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1rem;
	}
	h1 {
		margin: 0;
	}
	.sub {
		margin: 0.25rem 0 0;
		font-size: 0.92rem;
	}
	.instructions {
		margin: 0 0 1rem;
		display: flex;
		align-items: center;
		gap: 0.4rem;
		flex-wrap: wrap;
	}
	.small {
		font-size: 0.85rem;
	}
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		padding: 2.5rem;
		text-align: center;
		color: var(--muted);
		margin-bottom: 1rem;
	}
	.empty .emoji {
		font-size: 2.5rem;
	}
</style>
