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

<div class="head">
	<div>
		<h1>My Map</h1>
		<p class="muted sub">Click a country to mark where you've cracked nuts with someone.</p>
	</div>
	{#if user.id}
		<div class="stat">
			<span class="num">{total}</span>
			<span class="muted">cracked</span>
		</div>
	{/if}
</div>

{#if !user.id}
	<div class="card empty">
		<span class="emoji">🌍</span>
		<p>Enter your user ID in the top-right to load your map.</p>
	</div>
{:else}
	<div class="card controls">
		<label class="switch">
			<input type="checkbox" checked={countMode} onchange={toggleCountMode} />
			<span class="track"><span class="thumb"></span></span>
			<span class="switch-label">
				Count mode
				<span class="muted small">
					{countMode ? 'left-click adds a crack' : 'left-click marks a country'}
				</span>
			</span>
		</label>
		<span class="muted small hint">Right-click a country to remove one</span>
		{#if loading}<span class="spinner"></span>{/if}
	</div>
{/if}

{#if error}<p class="error">⚠️ {error}</p>{/if}

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
	.controls {
		display: flex;
		align-items: center;
		gap: 1.25rem;
		padding: 0.85rem 1.1rem;
		margin-bottom: 1rem;
		flex-wrap: wrap;
	}
	.small {
		font-size: 0.8rem;
	}
	.hint {
		margin-left: auto;
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
	.error {
		color: var(--danger);
		font-size: 0.9rem;
	}

	/* toggle switch */
	.switch {
		display: inline-flex;
		align-items: center;
		gap: 0.65rem;
		cursor: pointer;
		user-select: none;
	}
	.switch input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}
	.track {
		position: relative;
		width: 42px;
		height: 24px;
		border-radius: 999px;
		background: var(--surface-2);
		border: 1px solid var(--border);
		transition: background 0.2s var(--ease);
		flex-shrink: 0;
	}
	.thumb {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: white;
		box-shadow: var(--shadow-sm);
		transition: transform 0.2s var(--ease);
	}
	.switch input:checked + .track {
		background: var(--primary);
		border-color: var(--primary);
	}
	.switch input:checked + .track .thumb {
		transform: translateX(18px);
	}
	.switch input:focus-visible + .track {
		box-shadow: var(--ring);
	}
	.switch-label {
		display: flex;
		flex-direction: column;
		font-size: 0.9rem;
		font-weight: 500;
		line-height: 1.2;
	}
</style>
