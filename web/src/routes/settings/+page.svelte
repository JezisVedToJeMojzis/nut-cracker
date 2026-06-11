<script lang="ts">
	import { getSettings, updateSettings } from '$lib/api';
	import { user } from '$lib/user.svelte';

	let countMode = $state(false);
	let loaded = $state(false);
	let error = $state('');

	$effect(() => {
		if (user.id) load(user.id);
	});

	async function load(id: string) {
		error = '';
		try {
			const st = await getSettings(id);
			countMode = st.count_mode;
			loaded = true;
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		}
	}

	async function toggleCountMode() {
		const next = !countMode;
		countMode = next; // optimistic
		try {
			const st = await updateSettings(user.id, { count_mode: next });
			countMode = st.count_mode;
		} catch (e) {
			countMode = !next;
			error = e instanceof Error ? e.message : String(e);
		}
	}
</script>

<h1>Settings</h1>
<p class="muted sub">Enable or disable features for your map.</p>

{#if !user.id}
	<div class="card empty">Enter your user ID in the top-right to manage settings.</div>
{:else}
	{#if error}<p class="error">⚠️ {error}</p>{/if}

	<div class="card feature">
		<label class="switch">
			<input
				type="checkbox"
				checked={countMode}
				disabled={!loaded}
				onchange={toggleCountMode}
			/>
			<span class="track"><span class="thumb"></span></span>
		</label>
		<div class="info">
			<span class="title">Count mode</span>
			<span class="muted small">
				Track how many people you've cracked nuts with per country. When off, a country is simply
				marked; when on, each click adds to the count and the colour deepens.
			</span>
		</div>
	</div>
{/if}

<style>
	h1 {
		margin: 0;
	}
	.sub {
		margin: 0.25rem 0 1.25rem;
		font-size: 0.92rem;
	}
	.feature {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		padding: 1.1rem 1.25rem;
		max-width: 560px;
	}
	.info {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}
	.title {
		font-weight: 600;
	}
	.small {
		font-size: 0.85rem;
		line-height: 1.4;
	}
	.empty {
		padding: 1.5rem;
		text-align: center;
		color: var(--muted);
	}
	.error {
		color: var(--danger);
		font-size: 0.9rem;
	}

	/* toggle switch */
	.switch {
		display: inline-flex;
		cursor: pointer;
		flex-shrink: 0;
		margin-top: 0.15rem;
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
</style>
