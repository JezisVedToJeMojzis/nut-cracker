<script lang="ts">
	import { onMount } from 'svelte';
	import WorldMap from '$lib/WorldMap.svelte';
	import { getMap, increment, decrement } from '$lib/api';

	let userId = $state('');
	let countMode = $state(false);
	let cracks = $state<Record<string, number>>({});
	let error = $state('');
	let loading = $state(false);

	onMount(() => {
		const saved = localStorage.getItem('nutcracker_user_id');
		if (saved) {
			userId = saved;
			loadMap();
		}
	});

	async function loadMap() {
		if (!userId) return;
		loading = true;
		error = '';
		try {
			const list = await getMap(userId, userId);
			const next: Record<string, number> = {};
			for (const c of list) next[c.country_code] = c.cracks;
			cracks = next;
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	}

	function saveUser() {
		localStorage.setItem('nutcracker_user_id', userId);
		loadMap();
	}

	async function onCrack(code: string) {
		if (!userId) {
			error = 'Set your user ID first';
			return;
		}
		try {
			const res = await increment(userId, code);
			cracks = { ...cracks, [code]: res.cracks };
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		}
	}

	async function onUncrack(code: string) {
		if (!userId || !(code in cracks)) return;
		try {
			const res = await decrement(userId, code);
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

<main>
	<h1>🥜 Nut Cracker</h1>
	<p class="tagline">Color the countries where you've cracked nuts with someone.</p>

	<div class="controls">
		<label>
			Your user ID
			<input
				type="text"
				bind:value={userId}
				placeholder="paste a user UUID"
				onkeydown={(e) => e.key === 'Enter' && saveUser()}
			/>
		</label>
		<button onclick={saveUser}>Load my map</button>

		<label class="toggle">
			<input type="checkbox" bind:checked={countMode} />
			Count mode
		</label>
	</div>

	<p class="hint">
		Left-click a country to {countMode ? 'add a crack' : 'mark it'}. Right-click to remove one.
		{#if total > 0}<strong>{total}</strong> countries cracked.{/if}
	</p>

	{#if error}
		<p class="error">⚠️ {error}</p>
	{/if}
	{#if loading}
		<p>Loading…</p>
	{/if}

	<WorldMap {cracks} {countMode} oncrack={onCrack} onuncrack={onUncrack} />
</main>

<style>
	main {
		max-width: 1100px;
		margin: 0 auto;
		padding: 1.5rem 1rem 3rem;
		font-family: system-ui, sans-serif;
		color: #111827;
	}
	h1 {
		margin-bottom: 0.25rem;
	}
	.tagline {
		margin-top: 0;
		color: #6b7280;
	}
	.controls {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: flex-end;
		margin: 1rem 0;
	}
	label {
		display: flex;
		flex-direction: column;
		font-size: 0.85rem;
		gap: 0.25rem;
	}
	.toggle {
		flex-direction: row;
		align-items: center;
		gap: 0.4rem;
		font-size: 1rem;
	}
	input[type='text'] {
		padding: 0.4rem 0.6rem;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		min-width: 320px;
	}
	button {
		padding: 0.45rem 0.9rem;
		border: none;
		background: #16a34a;
		color: white;
		border-radius: 6px;
		cursor: pointer;
	}
	button:hover {
		background: #15803d;
	}
	.hint {
		color: #6b7280;
		font-size: 0.9rem;
	}
	.error {
		color: #b91c1c;
	}
</style>
