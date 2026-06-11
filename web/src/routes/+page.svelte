<script lang="ts">
	import WorldMap from '$lib/WorldMap.svelte';
	import { getMap, increment, decrement, getSettings, updateSettings } from '$lib/api';
	import { user } from '$lib/user.svelte';

	let countMode = $state(false);
	let cracks = $state<Record<string, number>>({});
	let error = $state('');
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
		error = '';
		try {
			const [list, st] = await Promise.all([getMap(id, id), getSettings(id)]);
			const next: Record<string, number> = {};
			for (const c of list) next[c.country_code] = c.cracks;
			cracks = next;
			countMode = st.count_mode;
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	async function toggleCountMode() {
		if (!user.id) {
			error = 'Set your user ID first';
			return;
		}
		const next = !countMode;
		countMode = next; // optimistic
		try {
			const st = await updateSettings(user.id, { count_mode: next });
			countMode = st.count_mode;
		} catch (e) {
			countMode = !next; // revert on failure
			error = e instanceof Error ? e.message : String(e);
		}
	}

	async function onCrack(code: string) {
		if (!user.id) {
			error = 'Set your user ID first';
			return;
		}
		// With count mode off, an already-coloured country stays as-is.
		if (!countMode && code in cracks) return;
		try {
			const res = await increment(user.id, code);
			cracks = { ...cracks, [code]: res.cracks };
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
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
			error = e instanceof Error ? e.message : String(e);
		}
	}

	const total = $derived(Object.keys(cracks).length);
</script>

<h1>My Map</h1>

{#if !user.id}
	<p class="hint">Enter your user ID in the top-right to load your map.</p>
{:else}
	<div class="controls">
		<label class="toggle">
			<input type="checkbox" checked={countMode} onchange={toggleCountMode} />
			Count mode
		</label>
		<span class="hint">
			Left-click to {countMode ? 'add a crack' : 'mark a country'}. Right-click to remove one.
			{#if total > 0}<strong>{total}</strong> cracked.{/if}
		</span>
	</div>
{/if}

{#if error}<p class="error">⚠️ {error}</p>{/if}
{#if loading}<p>Loading…</p>{/if}

<WorldMap {cracks} {countMode} oncrack={onCrack} onuncrack={onUncrack} />

<style>
	h1 {
		margin-top: 0;
	}
	.controls {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin: 0.5rem 0 1rem;
		flex-wrap: wrap;
	}
	.toggle {
		display: flex;
		align-items: center;
		gap: 0.4rem;
	}
	.hint {
		color: #6b7280;
		font-size: 0.9rem;
	}
	.error {
		color: #b91c1c;
	}
</style>
