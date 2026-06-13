<script lang="ts">
	import world from '@svg-maps/world';

	type Props = {
		cracks: Record<string, number>;
		countMode: boolean;
	};
	let { cracks, countMode }: Props = $props();

	// Build an UPPERCASE ISO code -> country name lookup once.
	const names: Record<string, string> = {};
	for (const l of world.locations as { id: string; name: string }[]) {
		names[l.id.toUpperCase()] = l.name;
	}
	const totalCountriesOnMap = world.locations.length;

	const countries = $derived(Object.keys(cracks).length);
	const totalCracks = $derived(Object.values(cracks).reduce((a, b) => a + b, 0));
	const worldPct = $derived(Math.round((countries / totalCountriesOnMap) * 100));
	const top = $derived(
		Object.entries(cracks)
			.sort((a, b) => b[1] - a[1])
			.slice(0, 5)
			.map(([code, count]) => ({ code, count, name: names[code] ?? code }))
	);
</script>

<div class="stats">
	<div class="cards">
		<div class="stat-card">
			<span class="num">{countries}</span>
			<span class="label">Countries</span>
		</div>
		{#if countMode}
			<div class="stat-card">
				<span class="num">{totalCracks}</span>
				<span class="label">Total cracks</span>
			</div>
		{/if}
		<div class="stat-card">
			<span class="num">{worldPct}%</span>
			<span class="label">of the world</span>
		</div>
	</div>

	{#if countMode && top.length > 0}
		<div class="top">
			<span class="top-title">Top countries</span>
			<div class="chips">
				{#each top as t (t.code)}
					<span class="chip">
						{t.name}
						<span class="cnt">{t.count}</span>
					</span>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.stats {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: stretch;
		margin: 0 0 1rem;
	}
	.cards {
		display: flex;
		gap: 0.6rem;
	}
	.stat-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-width: 92px;
		padding: 0.6rem 0.9rem;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		box-shadow: var(--shadow-sm);
		line-height: 1.15;
	}
	.stat-card .num {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--primary-700);
	}
	.stat-card .label {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--muted);
	}
	.top {
		flex: 1;
		min-width: 200px;
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
		justify-content: center;
		padding: 0.6rem 0.9rem;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		box-shadow: var(--shadow-sm);
	}
	.top-title {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--muted);
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
	}
	.chip {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.2rem 0.6rem;
		border-radius: 999px;
		background: var(--surface-2);
		font-size: 0.82rem;
	}
	.cnt {
		display: grid;
		place-items: center;
		min-width: 18px;
		height: 18px;
		padding: 0 4px;
		border-radius: 999px;
		background: var(--primary);
		color: white;
		font-size: 0.7rem;
		font-weight: 700;
	}
</style>
