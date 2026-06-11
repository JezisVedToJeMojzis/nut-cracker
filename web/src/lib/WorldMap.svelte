<script lang="ts">
	import world from '@svg-maps/world';

	type Props = {
		/** Map of UPPERCASE ISO code -> crack count. */
		cracks: Record<string, number>;
		/** Whether to show the numeric count badge / intensity by count. */
		countMode: boolean;
		/** Left-click a country. */
		oncrack: (code: string) => void;
		/** Right-click a country. */
		onuncrack: (code: string) => void;
	};

	let { cracks, countMode, oncrack, onuncrack }: Props = $props();

	type Location = { name: string; id: string; path: string };

	// @svg-maps/world ids are lowercase ISO-2; our backend uses uppercase.
	const locations = (world.locations as Location[]).map((l) => ({
		...l,
		code: l.id.toUpperCase()
	}));

	let hovered = $state<{ name: string; code: string; count: number } | null>(null);

	// Fill color: uncracked = light grey; cracked = green, deeper with more
	// cracks when count mode is on.
	function fillFor(code: string): string {
		const count = cracks[code] ?? 0;
		if (count <= 0) return '#dfe6ee';
		if (!countMode) return '#10b981';
		// Count mode: ramp from light to dark green over counts 1..6+.
		const step = Math.min(count, 6);
		const lightness = 58 - step * 6; // 52% .. 22%
		return `hsl(160, 70%, ${lightness}%)`;
	}

	function handleClick(e: MouseEvent, code: string) {
		e.preventDefault();
		oncrack(code);
	}

	function handleContext(e: MouseEvent, code: string) {
		e.preventDefault();
		onuncrack(code);
	}
</script>

<div class="map-wrap">
	<svg viewBox={world.viewBox} xmlns="http://www.w3.org/2000/svg" role="group" aria-label="World map">
		{#each locations as loc (loc.code)}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<path
				d={loc.path}
				class="country"
				fill={fillFor(loc.code)}
				role="button"
				tabindex="-1"
				aria-label={loc.name}
				onclick={(e) => handleClick(e, loc.code)}
				oncontextmenu={(e) => handleContext(e, loc.code)}
				onmouseenter={() =>
					(hovered = { name: loc.name, code: loc.code, count: cracks[loc.code] ?? 0 })}
				onmouseleave={() => (hovered = null)}
			/>
		{/each}
	</svg>

	{#if hovered}
		<div class="tooltip">
			<strong>{hovered.name}</strong>
			{#if hovered.count > 0}
				<span class="count">{hovered.count} crack{hovered.count === 1 ? '' : 's'}</span>
			{/if}
		</div>
	{/if}
</div>

<style>
	.map-wrap {
		position: relative;
		width: 100%;
	}
	svg {
		width: 100%;
		height: auto;
		background: linear-gradient(160deg, #dbeafe 0%, #c7e7f5 50%, #bfe9e4 100%);
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		border: 1px solid var(--border);
	}
	.country {
		stroke: rgba(255, 255, 255, 0.85);
		stroke-width: 0.4;
		stroke-linejoin: round;
		cursor: pointer;
		transition:
			fill 0.25s var(--ease),
			filter 0.15s var(--ease);
	}
	.country:hover {
		stroke: #0f172a;
		stroke-width: 0.9;
		filter: brightness(1.08) drop-shadow(0 1px 2px rgba(0, 0, 0, 0.25));
	}
	.tooltip {
		position: absolute;
		top: 12px;
		left: 12px;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: rgba(15, 23, 42, 0.9);
		backdrop-filter: blur(6px);
		color: white;
		padding: 0.4rem 0.75rem;
		border-radius: 999px;
		font-size: 0.85rem;
		pointer-events: none;
		box-shadow: var(--shadow-lg);
		animation: tip-in 0.15s var(--ease);
	}
	.count {
		background: rgba(255, 255, 255, 0.18);
		padding: 0.05rem 0.45rem;
		border-radius: 999px;
		font-size: 0.78rem;
		font-weight: 600;
	}
	@keyframes tip-in {
		from {
			opacity: 0;
			transform: translateY(-4px);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
</style>
